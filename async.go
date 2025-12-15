package birdactyl

import (
	"context"
	"encoding/json"

	pb "github.com/pizzlad/birdactyl-go-sdk/proto"
	"google.golang.org/grpc/metadata"
)

type AsyncAPI struct {
	panel    pb.PanelServiceClient
	pluginID string
}

func (a *AsyncAPI) ctx() context.Context {
	return metadata.AppendToOutgoingContext(context.Background(), "x-plugin-id", a.pluginID)
}

type Future[T any] struct {
	ch  chan result[T]
	val *T
	err error
}

type result[T any] struct {
	val T
	err error
}

func newFuture[T any](fn func() (T, error)) *Future[T] {
	f := &Future[T]{ch: make(chan result[T], 1)}
	go func() {
		val, err := fn()
		f.ch <- result[T]{val: val, err: err}
	}()
	return f
}

func (f *Future[T]) Get() (T, error) {
	if f.val != nil || f.err != nil {
		return *f.val, f.err
	}
	r := <-f.ch
	f.val = &r.val
	f.err = r.err
	return r.val, r.err
}

func (f *Future[T]) Then(fn func(T)) *Future[T] {
	go func() {
		val, err := f.Get()
		if err == nil {
			fn(val)
		}
	}()
	return f
}

func (f *Future[T]) Catch(fn func(error)) *Future[T] {
	go func() {
		_, err := f.Get()
		if err != nil {
			fn(err)
		}
	}()
	return f
}

func (a *AsyncAPI) GetServer(id string) *Future[*Server] {
	return newFuture(func() (*Server, error) {
		r, err := a.panel.GetServer(a.ctx(), &pb.IDRequest{Id: id})
		if err != nil {
			return nil, err
		}
		return &Server{ID: r.Id, Name: r.Name, OwnerID: r.UserId, NodeID: r.NodeId, Status: r.Status, Suspended: r.Suspended, Memory: r.Memory, Disk: r.Disk, CPU: r.Cpu, PackageID: r.PackageId, PrimaryAllocation: r.PrimaryAllocation}, nil
	})
}

func (a *AsyncAPI) ListServers() *Future[[]*Server] {
	return newFuture(func() ([]*Server, error) {
		r, err := a.panel.ListServers(a.ctx(), &pb.ListServersRequest{})
		if err != nil {
			return nil, err
		}
		out := make([]*Server, len(r.GetServers()))
		for i, s := range r.GetServers() {
			out[i] = &Server{ID: s.Id, Name: s.Name, OwnerID: s.UserId, NodeID: s.NodeId, Status: s.Status, Suspended: s.Suspended, Memory: s.Memory, Disk: s.Disk, CPU: s.Cpu, PackageID: s.PackageId, PrimaryAllocation: s.PrimaryAllocation}
		}
		return out, nil
	})
}

func (a *AsyncAPI) StartServer(id string) *Future[struct{}] {
	return newFuture(func() (struct{}, error) {
		_, err := a.panel.StartServer(a.ctx(), &pb.IDRequest{Id: id})
		return struct{}{}, err
	})
}

func (a *AsyncAPI) StopServer(id string) *Future[struct{}] {
	return newFuture(func() (struct{}, error) {
		_, err := a.panel.StopServer(a.ctx(), &pb.IDRequest{Id: id})
		return struct{}{}, err
	})
}

func (a *AsyncAPI) RestartServer(id string) *Future[struct{}] {
	return newFuture(func() (struct{}, error) {
		_, err := a.panel.RestartServer(a.ctx(), &pb.IDRequest{Id: id})
		return struct{}{}, err
	})
}

func (a *AsyncAPI) KillServer(id string) *Future[struct{}] {
	return newFuture(func() (struct{}, error) {
		_, err := a.panel.KillServer(a.ctx(), &pb.IDRequest{Id: id})
		return struct{}{}, err
	})
}

func (a *AsyncAPI) DeleteServer(id string) *Future[struct{}] {
	return newFuture(func() (struct{}, error) {
		_, err := a.panel.DeleteServer(a.ctx(), &pb.IDRequest{Id: id})
		return struct{}{}, err
	})
}

func (a *AsyncAPI) GetUser(id string) *Future[*User] {
	return newFuture(func() (*User, error) {
		r, err := a.panel.GetUser(a.ctx(), &pb.IDRequest{Id: id})
		if err != nil {
			return nil, err
		}
		return &User{ID: r.Id, Username: r.Username, Email: r.Email, IsAdmin: r.IsAdmin, IsBanned: r.IsBanned, ForcePasswordReset: r.ForcePasswordReset, RamLimit: r.RamLimit, CpuLimit: r.CpuLimit, DiskLimit: r.DiskLimit, ServerLimit: r.ServerLimit, CreatedAt: r.CreatedAt}, nil
	})
}

func (a *AsyncAPI) ListUsers() *Future[[]*User] {
	return newFuture(func() ([]*User, error) {
		r, err := a.panel.ListUsers(a.ctx(), &pb.ListUsersRequest{})
		if err != nil {
			return nil, err
		}
		out := make([]*User, len(r.GetUsers()))
		for i, u := range r.GetUsers() {
			out[i] = &User{ID: u.Id, Username: u.Username, Email: u.Email, IsAdmin: u.IsAdmin, IsBanned: u.IsBanned, ForcePasswordReset: u.ForcePasswordReset, RamLimit: u.RamLimit, CpuLimit: u.CpuLimit, DiskLimit: u.DiskLimit, ServerLimit: u.ServerLimit, CreatedAt: u.CreatedAt}
		}
		return out, nil
	})
}

func (a *AsyncAPI) GetNode(id string) *Future[*Node] {
	return newFuture(func() (*Node, error) {
		r, err := a.panel.GetNode(a.ctx(), &pb.IDRequest{Id: id})
		if err != nil {
			return nil, err
		}
		return &Node{ID: r.Id, Name: r.Name, FQDN: r.Fqdn, Port: r.Port, IsOnline: r.IsOnline, LastHeartbeat: r.LastHeartbeat}, nil
	})
}

func (a *AsyncAPI) ListNodes() *Future[[]*Node] {
	return newFuture(func() ([]*Node, error) {
		r, err := a.panel.ListNodes(a.ctx(), &pb.Empty{})
		if err != nil {
			return nil, err
		}
		out := make([]*Node, len(r.GetNodes()))
		for i, n := range r.GetNodes() {
			out[i] = &Node{ID: n.Id, Name: n.Name, FQDN: n.Fqdn, Port: n.Port, IsOnline: n.IsOnline, LastHeartbeat: n.LastHeartbeat}
		}
		return out, nil
	})
}

func (a *AsyncAPI) GetConsoleLog(serverID string, lines int32) *Future[[]string] {
	return newFuture(func() ([]string, error) {
		r, err := a.panel.GetConsoleLog(a.ctx(), &pb.ConsoleLogRequest{ServerId: serverID, Lines: lines})
		if err != nil {
			return nil, err
		}
		return r.Lines, nil
	})
}

func (a *AsyncAPI) SendCommand(serverID, command string) *Future[struct{}] {
	return newFuture(func() (struct{}, error) {
		_, err := a.panel.SendCommand(a.ctx(), &pb.SendCommandRequest{ServerId: serverID, Command: command})
		return struct{}{}, err
	})
}

func (a *AsyncAPI) GetServerStats(serverID string) *Future[*ServerStats] {
	return newFuture(func() (*ServerStats, error) {
		r, err := a.panel.GetServerStats(a.ctx(), &pb.IDRequest{Id: serverID})
		if err != nil {
			return nil, err
		}
		return &ServerStats{MemoryBytes: r.MemoryBytes, MemoryLimit: r.MemoryLimit, CPUPercent: r.CpuPercent, DiskBytes: r.DiskBytes, NetworkRx: r.NetworkRx, NetworkTx: r.NetworkTx, State: r.State}, nil
	})
}

func (a *AsyncAPI) GetFullLog(serverID string) *Future[[]byte] {
	return newFuture(func() ([]byte, error) {
		r, err := a.panel.GetFullLog(a.ctx(), &pb.IDRequest{Id: serverID})
		if err != nil {
			return nil, err
		}
		return r.Content, nil
	})
}

func (a *AsyncAPI) SearchLogs(serverID, pattern string, regex bool, limit int32) *Future[[]*LogMatch] {
	return newFuture(func() ([]*LogMatch, error) {
		r, err := a.panel.SearchLogs(a.ctx(), &pb.SearchLogsRequest{ServerId: serverID, Pattern: pattern, Regex: regex, Limit: limit})
		if err != nil {
			return nil, err
		}
		out := make([]*LogMatch, len(r.Matches))
		for i, m := range r.Matches {
			out[i] = &LogMatch{Line: m.Line, LineNumber: m.LineNumber, Timestamp: m.Timestamp}
		}
		return out, nil
	})
}

func (a *AsyncAPI) ListLogFiles(serverID string) *Future[[]*LogFile] {
	return newFuture(func() ([]*LogFile, error) {
		r, err := a.panel.ListLogFiles(a.ctx(), &pb.IDRequest{Id: serverID})
		if err != nil {
			return nil, err
		}
		out := make([]*LogFile, len(r.Files))
		for i, f := range r.Files {
			out[i] = &LogFile{Name: f.Name, Size: f.Size, Modified: f.Modified}
		}
		return out, nil
	})
}

func (a *AsyncAPI) ReadLogFile(serverID, filename string) *Future[[]byte] {
	return newFuture(func() ([]byte, error) {
		r, err := a.panel.ReadLogFile(a.ctx(), &pb.ReadLogFileRequest{ServerId: serverID, Filename: filename})
		if err != nil {
			return nil, err
		}
		return r.Content, nil
	})
}

func (a *AsyncAPI) ListFiles(serverID, path string) *Future[[]*File] {
	return newFuture(func() ([]*File, error) {
		r, err := a.panel.ListFiles(a.ctx(), &pb.FilePathRequest{ServerId: serverID, Path: path})
		if err != nil {
			return nil, err
		}
		out := make([]*File, len(r.GetFiles()))
		for i, f := range r.GetFiles() {
			out[i] = &File{Name: f.Name, Size: f.Size, IsDir: f.IsDir, ModTime: f.Modified, Mime: f.Mime}
		}
		return out, nil
	})
}

func (a *AsyncAPI) ReadFile(serverID, path string) *Future[[]byte] {
	return newFuture(func() ([]byte, error) {
		r, err := a.panel.ReadFile(a.ctx(), &pb.FilePathRequest{ServerId: serverID, Path: path})
		if err != nil {
			return nil, err
		}
		return r.Content, nil
	})
}

func (a *AsyncAPI) WriteFile(serverID, path string, content []byte) *Future[struct{}] {
	return newFuture(func() (struct{}, error) {
		_, err := a.panel.WriteFile(a.ctx(), &pb.WriteFileRequest{ServerId: serverID, Path: path, Content: content})
		return struct{}{}, err
	})
}

func (a *AsyncAPI) GetKV(key string) *Future[string] {
	return newFuture(func() (string, error) {
		r, err := a.panel.GetKV(a.ctx(), &pb.KVRequest{Key: key})
		if err != nil {
			return "", err
		}
		if !r.Found {
			return "", nil
		}
		return r.Value, nil
	})
}

func (a *AsyncAPI) SetKV(key, value string) *Future[struct{}] {
	return newFuture(func() (struct{}, error) {
		_, err := a.panel.SetKV(a.ctx(), &pb.KVSetRequest{Key: key, Value: value})
		return struct{}{}, err
	})
}

func (a *AsyncAPI) QueryDB(query string, args ...string) *Future[[]map[string]interface{}] {
	return newFuture(func() ([]map[string]interface{}, error) {
		r, err := a.panel.QueryDB(a.ctx(), &pb.QueryDBRequest{Query: query, Args: args})
		if err != nil {
			return nil, err
		}
		out := make([]map[string]interface{}, len(r.Rows))
		for i, row := range r.Rows {
			var m map[string]interface{}
			json.Unmarshal(row, &m)
			out[i] = m
		}
		return out, nil
	})
}

func (a *AsyncAPI) HTTP(method, url string, headers map[string]string, body []byte) *Future[*HTTPResponse] {
	return newFuture(func() (*HTTPResponse, error) {
		r, err := a.panel.HTTPRequest(a.ctx(), &pb.PluginHTTPRequest{Method: method, Url: url, Headers: headers, Body: body})
		if err != nil {
			return nil, err
		}
		return &HTTPResponse{Status: int(r.Status), Headers: r.Headers, Body: r.Body, Error: r.Error}, nil
	})
}

func All[T any](futures ...*Future[T]) *Future[[]T] {
	return newFuture(func() ([]T, error) {
		results := make([]T, len(futures))
		for i, f := range futures {
			val, err := f.Get()
			if err != nil {
				return nil, err
			}
			results[i] = val
		}
		return results, nil
	})
}
