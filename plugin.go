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

	pb "github.com/pizzlad/birdactyl-go-sdk/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type Plugin struct {
	id         string
	name       string
	version    string
	events     map[string]EventHandler
	routes     map[string]RouteHandler
	schedule   map[string]ScheduleHandler
	mixins     []MixinRegistration
	panel      pb.PanelServiceClient
	conn       *grpc.ClientConn
	api        *API
	dataDir    string
	useDataDir bool
	onStart    func()
	ui         *UIConfig
}

type UIConfig struct {
	Icon    string
	Sidebar *SidebarConfig
	Pages   []PageConfig
}

type SidebarConfig struct {
	Label   string
	Icon    string
	Section string
	Order   int
}

type PageConfig struct {
	Path  string
	Label string
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
		mixins:   make([]MixinRegistration, 0),
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

func (p *Plugin) UI(cfg UIConfig) *Plugin {
	p.ui = &cfg
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

func (p *Plugin) Mixin(target string, handler MixinHandler) *Plugin {
	return p.MixinWithPriority(target, 0, handler)
}

func (p *Plugin) MixinWithPriority(target string, priority int, handler MixinHandler) *Plugin {
	p.mixins = append(p.mixins, MixinRegistration{
		Target:   target,
		Priority: priority,
		Handler:  handler,
	})
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

	if p.onStart != nil {
		p.onStart()
	}

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

	mixins := make([]*pb.MixinInfo, 0, len(s.plugin.mixins))
	for _, m := range s.plugin.mixins {
		mixins = append(mixins, &pb.MixinInfo{Target: m.Target, Priority: int32(m.Priority)})
	}

	info := &pb.PluginInfo{
		Id:        s.plugin.id,
		Name:      s.plugin.name,
		Version:   s.plugin.version,
		Events:    events,
		Routes:    routes,
		Schedules: schedules,
		Mixins:    mixins,
	}

	if s.plugin.ui != nil {
		info.Ui = &pb.PluginUIInfo{Icon: s.plugin.ui.Icon}
		if s.plugin.ui.Sidebar != nil {
			info.Ui.Sidebar = &pb.PluginSidebarInfo{
				Label:   s.plugin.ui.Sidebar.Label,
				Icon:    s.plugin.ui.Sidebar.Icon,
				Section: s.plugin.ui.Sidebar.Section,
				Order:   int32(s.plugin.ui.Sidebar.Order),
			}
		}
		for _, p := range s.plugin.ui.Pages {
			info.Ui.Pages = append(info.Ui.Pages, &pb.PluginPageInfo{Path: p.Path, Label: p.Label})
		}
	}

	return info, nil
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

func (s *pluginServer) OnMixin(ctx context.Context, req *pb.MixinRequest) (*pb.MixinResponse, error) {
	var handler MixinHandler
	for _, m := range s.plugin.mixins {
		if m.Target == req.Target {
			handler = m.Handler
			break
		}
	}

	if handler == nil {
		return &pb.MixinResponse{Action: pb.MixinResponse_NEXT}, nil
	}

	var input map[string]interface{}
	json.Unmarshal(req.Input, &input)

	var chainData map[string]interface{}
	if len(req.ChainData) > 0 {
		json.Unmarshal(req.ChainData, &chainData)
	}

	mctx := &MixinContext{
		Target:    req.Target,
		RequestID: req.RequestId,
		input:     input,
		chainData: chainData,
	}

	result := handler(mctx)

	resp := &pb.MixinResponse{
		Action: pb.MixinResponse_Action(result.action),
	}

	if result.output != nil {
		resp.Output, _ = json.Marshal(result.output)
	}
	if result.err != "" {
		resp.Error = result.err
	}
	if result.modifiedInput != nil {
		resp.ModifiedInput, _ = json.Marshal(result.modifiedInput)
	}

	return resp, nil
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
