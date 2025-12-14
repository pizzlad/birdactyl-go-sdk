package birdactyl

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"

	pb "github.com/pizzlad/birdactyl-go-sdk/internal/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type Plugin struct {
	id          string
	name        string
	version     string
	events      map[string]EventHandler
	routes      map[string]RouteHandler
	schedule    map[string]ScheduleHandler
	panel       pb.PanelServiceClient
	conn        *grpc.ClientConn
	api         *API
	dataDir     string
	useDataDir  bool
	onStart     func()
}

type EventHandler func(Event) EventResult
type RouteHandler func(Request) Response
type ScheduleHandler func()

func New(id, version string) *Plugin {
	return &Plugin{
		id:       id,
		name:     id,
		version:  version,
		events:   make(map[string]EventHandler),
		routes:   make(map[string]RouteHandler),
		schedule: make(map[string]ScheduleHandler),
	}
}

func (p *Plugin) SetName(name string) *Plugin {
	p.name = name
	return p
}

func (p *Plugin) UseDataDir() *Plugin {
	p.useDataDir = true
	return p
}

func (p *Plugin) OnStart(fn func()) *Plugin {
	p.onStart = fn
	return p
}

func (p *Plugin) OnEvent(eventType string, handler EventHandler) *Plugin {
	p.events[eventType] = handler
	return p
}

func (p *Plugin) Route(method, path string, handler RouteHandler) *Plugin {
	p.routes[method+":"+path] = handler
	return p
}

func (p *Plugin) Schedule(id, cron string, handler ScheduleHandler) *Plugin {
	p.schedule[id+":"+cron] = handler
	return p
}

func (p *Plugin) API() *API {
	return p.api
}

func (p *Plugin) Log(msg string) {
	ctx := metadata.AppendToOutgoingContext(context.Background(), "x-plugin-id", p.id)
	p.panel.Log(ctx, &pb.LogRequest{Level: "info", Message: msg})
}

func (p *Plugin) DataDir() string {
	return p.dataDir
}

func (p *Plugin) DataPath(filename string) string {
	return filepath.Join(p.dataDir, filename)
}

func (p *Plugin) SaveConfig(v interface{}) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p.DataPath("config.json"), data, 0644)
}

func (p *Plugin) LoadConfig(v interface{}) error {
	data, err := os.ReadFile(p.DataPath("config.json"))
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

func (p *Plugin) Start(panelAddr string, defaultPort int) error {
	port := defaultPort
	if len(os.Args) > 1 {
		if pt, err := strconv.Atoi(os.Args[1]); err == nil {
			port = pt
		}
	}

	if len(os.Args) > 2 {
		p.dataDir = filepath.Join(os.Args[2], p.id+"_data")
	} else {
		p.dataDir = p.id + "_data"
	}

	if p.useDataDir {
		if err := os.MkdirAll(p.dataDir, 0755); err != nil {
			log.Printf("[%s] failed to create data dir %s: %v", p.id, p.dataDir, err)
		}
	}

	if p.onStart != nil {
		p.onStart()
	}

	conn, err := grpc.NewClient(panelAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
			ctx = metadata.AppendToOutgoingContext(ctx, "x-plugin-id", p.id)
			return invoker(ctx, method, req, reply, cc, opts...)
		}),
	)
	if err != nil {
		return err
	}
	p.conn = conn
	p.panel = pb.NewPanelServiceClient(conn)
	p.api = &API{panel: p.panel, pluginID: p.id}

	p.Log(p.name + " v" + p.version + " started")

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}

	s := grpc.NewServer()
	pb.RegisterPluginServiceServer(s, &pluginServer{plugin: p})

	log.Printf("[%s] v%s listening on port %d", p.id, p.version, port)
	return s.Serve(lis)
}

type pluginServer struct {
	pb.UnimplementedPluginServiceServer
	plugin *Plugin
}

func (s *pluginServer) GetInfo(ctx context.Context, req *pb.Empty) (*pb.PluginInfo, error) {
	events := make([]string, 0, len(s.plugin.events))
	for e := range s.plugin.events {
		events = append(events, e)
	}

	routes := make([]*pb.RouteInfo, 0, len(s.plugin.routes))
	for key := range s.plugin.routes {
		method, path := splitKey(key)
		routes = append(routes, &pb.RouteInfo{Method: method, Path: path})
	}

	schedules := make([]*pb.ScheduleInfo, 0, len(s.plugin.schedule))
	for key := range s.plugin.schedule {
		id, cron := splitKey(key)
		schedules = append(schedules, &pb.ScheduleInfo{Id: id, Cron: cron})
	}

	return &pb.PluginInfo{
		Id:        s.plugin.id,
		Name:      s.plugin.name,
		Version:   s.plugin.version,
		Events:    events,
		Routes:    routes,
		Schedules: schedules,
	}, nil
}

func (s *pluginServer) OnEvent(ctx context.Context, ev *pb.Event) (*pb.EventResponse, error) {
	handler, ok := s.plugin.events[ev.Type]
	if !ok {
		return &pb.EventResponse{Allow: true}, nil
	}
	result := handler(Event{Type: ev.Type, Data: ev.Data, Sync: ev.Sync})
	return &pb.EventResponse{Allow: result.allow, Message: result.message}, nil
}

func (s *pluginServer) OnHTTP(ctx context.Context, req *pb.HTTPRequest) (*pb.HTTPResponse, error) {
	handler, ok := s.plugin.routes[req.Method+":"+req.Path]
	if !ok {
		for key, h := range s.plugin.routes {
			method, path := splitKey(key)
			if (method == "*" || method == req.Method) && matchPath(path, req.Path) {
				handler = h
				break
			}
		}
	}
	if handler == nil {
		return errorResponse(404, "not found"), nil
	}

	var body map[string]interface{}
	json.Unmarshal(req.Body, &body)

	resp := handler(Request{
		Method:  req.Method,
		Path:    req.Path,
		Headers: req.Headers,
		Query:   req.Query,
		Body:    body,
		RawBody: req.Body,
		UserID:  req.UserId,
	})

	return &pb.HTTPResponse{
		Status:  int32(resp.Status),
		Headers: resp.Headers,
		Body:    resp.body,
	}, nil
}

func (s *pluginServer) OnSchedule(ctx context.Context, req *pb.ScheduleRequest) (*pb.Empty, error) {
	for key, handler := range s.plugin.schedule {
		id, _ := splitKey(key)
		if id == req.ScheduleId {
			handler()
			break
		}
	}
	return &pb.Empty{}, nil
}

func (s *pluginServer) Shutdown(ctx context.Context, req *pb.Empty) (*pb.Empty, error) {
	log.Printf("[%s] shutdown", s.plugin.id)
	return &pb.Empty{}, nil
}

func splitKey(key string) (string, string) {
	for i, c := range key {
		if c == ':' {
			return key[:i], key[i+1:]
		}
	}
	return key, ""
}

func matchPath(pattern, path string) bool {
	if pattern == path {
		return true
	}
	if len(pattern) > 0 && pattern[len(pattern)-1] == '*' {
		return len(path) >= len(pattern)-1 && path[:len(pattern)-1] == pattern[:len(pattern)-1]
	}
	return false
}

func errorResponse(status int, msg string) *pb.HTTPResponse {
	b, _ := json.Marshal(map[string]interface{}{"success": false, "error": msg})
	return &pb.HTTPResponse{Status: int32(status), Headers: map[string]string{"Content-Type": "application/json"}, Body: b}
}
