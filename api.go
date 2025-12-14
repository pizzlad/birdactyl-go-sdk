package birdactyl

import (
	"context"

	pb "github.com/pizzlad/birdactyl-go-sdk/proto"
	"google.golang.org/grpc/metadata"
)

type API struct {
	panel    pb.PanelServiceClient
	pluginID string
}

func (a *API) ctx() context.Context {
	return metadata.AppendToOutgoingContext(context.Background(), "x-plugin-id", a.pluginID)
}

type Server struct {
	ID        string
	Name      string
	OwnerID   string
	NodeID    string
	Status    string
	Suspended bool
	Memory    int32
	Disk      int32
	CPU       int32
}

type User struct {
	ID       string
	Username string
	Email    string
	IsAdmin  bool
	IsBanned bool
}

type Node struct {
	ID       string
	Name     string
	FQDN     string
	Port     int32
	IsOnline bool
}

func (a *API) GetServer(id string) (*Server, error) {
	r, err := a.panel.GetServer(a.ctx(), &pb.IDRequest{Id: id})
	if err != nil {
		return nil, err
	}
	return serverFromProto(r), nil
}

func (a *API) ListServers() []*Server {
	r, _ := a.panel.ListServers(a.ctx(), &pb.ListServersRequest{})
	servers := make([]*Server, len(r.GetServers()))
	for i, s := range r.GetServers() {
		servers[i] = serverFromProto(s)
	}
	return servers
}

func (a *API) StartServer(id string) error {
	_, err := a.panel.StartServer(a.ctx(), &pb.IDRequest{Id: id})
	return err
}

func (a *API) StopServer(id string) error {
	_, err := a.panel.StopServer(a.ctx(), &pb.IDRequest{Id: id})
	return err
}

func (a *API) RestartServer(id string) error {
	_, err := a.panel.RestartServer(a.ctx(), &pb.IDRequest{Id: id})
	return err
}

func (a *API) KillServer(id string) error {
	_, err := a.panel.KillServer(a.ctx(), &pb.IDRequest{Id: id})
	return err
}

func (a *API) SuspendServer(id string) error {
	_, err := a.panel.SuspendServer(a.ctx(), &pb.IDRequest{Id: id})
	return err
}

func (a *API) UnsuspendServer(id string) error {
	_, err := a.panel.UnsuspendServer(a.ctx(), &pb.IDRequest{Id: id})
	return err
}

func (a *API) DeleteServer(id string) error {
	_, err := a.panel.DeleteServer(a.ctx(), &pb.IDRequest{Id: id})
	return err
}

func (a *API) GetUser(id string) (*User, error) {
	r, err := a.panel.GetUser(a.ctx(), &pb.IDRequest{Id: id})
	if err != nil {
		return nil, err
	}
	return userFromProto(r), nil
}

func (a *API) ListUsers() []*User {
	r, _ := a.panel.ListUsers(a.ctx(), &pb.ListUsersRequest{})
	users := make([]*User, len(r.GetUsers()))
	for i, u := range r.GetUsers() {
		users[i] = userFromProto(u)
	}
	return users
}

func (a *API) BanUser(id string) error {
	_, err := a.panel.BanUser(a.ctx(), &pb.IDRequest{Id: id})
	return err
}

func (a *API) UnbanUser(id string) error {
	_, err := a.panel.UnbanUser(a.ctx(), &pb.IDRequest{Id: id})
	return err
}

func (a *API) ListNodes() []*Node {
	r, _ := a.panel.ListNodes(a.ctx(), &pb.Empty{})
	nodes := make([]*Node, len(r.GetNodes()))
	for i, n := range r.GetNodes() {
		nodes[i] = &Node{ID: n.Id, Name: n.Name, FQDN: n.Fqdn, Port: n.Port, IsOnline: n.IsOnline}
	}
	return nodes
}

func (a *API) GetKV(key string) (string, bool) {
	r, _ := a.panel.GetKV(a.ctx(), &pb.KVRequest{Key: key})
	return r.GetValue(), r.GetFound()
}

func (a *API) SetKV(key, value string) {
	a.panel.SetKV(a.ctx(), &pb.KVSetRequest{Key: key, Value: value})
}

func (a *API) DeleteKV(key string) {
	a.panel.DeleteKV(a.ctx(), &pb.KVRequest{Key: key})
}

func (a *API) BroadcastEvent(eventType string, data map[string]string) {
	a.panel.BroadcastEvent(a.ctx(), &pb.BroadcastEventRequest{EventType: eventType, Data: data})
}

func serverFromProto(s *pb.Server) *Server {
	return &Server{
		ID: s.Id, Name: s.Name, OwnerID: s.UserId, NodeID: s.NodeId,
		Status: s.Status, Suspended: s.Suspended,
		Memory: s.Memory, Disk: s.Disk, CPU: s.Cpu,
	}
}

func userFromProto(u *pb.User) *User {
	return &User{ID: u.Id, Username: u.Username, Email: u.Email, IsAdmin: u.IsAdmin, IsBanned: u.IsBanned}
}
