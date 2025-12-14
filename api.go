package birdactyl

import (
	"context"
	"encoding/json"
	"fmt"

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

func (a *API) Log(level, message string) {
	a.panel.Log(a.ctx(), &pb.LogRequest{Level: level, Message: message})
}

type Server struct {
	ID                string
	Name              string
	OwnerID           string
	NodeID            string
	Status            string
	Suspended         bool
	Memory            int32
	Disk              int32
	CPU               int32
	PackageID         string
	PrimaryAllocation string
}

type User struct {
	ID                 string
	Username           string
	Email              string
	IsAdmin            bool
	IsBanned           bool
	ForcePasswordReset bool
	RamLimit           int32
	CpuLimit           int32
	DiskLimit          int32
	ServerLimit        int32
	CreatedAt          string
}

type Node struct {
	ID            string
	Name          string
	FQDN          string
	Port          int32
	IsOnline      bool
	LastHeartbeat string
}

type Database struct {
	ID       string
	Name     string
	Username string
	Host     string
	Port     int32
	Password string
}

type DatabaseHost struct {
	ID             string
	Name           string
	Host           string
	Port           int32
	Username       string
	MaxDatabases   int32
	DatabasesCount int32
}

type Backup struct {
	ID        string
	Name      string
	Size      int64
	CreatedAt string
}

type File struct {
	Name    string
	Size    int64
	IsDir   bool
	ModTime string
	Mime    string
}

type Package struct {
	ID             string
	Name           string
	Description    string
	DockerImage    string
	StartupCommand string
	StopCommand    string
	ConfigFiles    string
	Memory         int32
	CPU            int32
	Disk           int32
	IsPublic       bool
}

type IPBan struct {
	ID        string
	IP        string
	Reason    string
	CreatedAt string
}

type Subuser struct {
	ID          string
	UserID      string
	Username    string
	Email       string
	Permissions []string
}

type Settings struct {
	RegistrationEnabled     bool
	ServerCreationEnabled   bool
}

type ActivityLog struct {
	ID          string
	UserID      string
	Username    string
	Action      string
	Description string
	IP          string
	IsAdmin     bool
	CreatedAt   string
}

func (a *API) GetServer(id string) (*Server, error) {
	r, err := a.panel.GetServer(a.ctx(), &pb.IDRequest{Id: id})
	if err != nil {
		return nil, err
	}
	return &Server{ID: r.Id, Name: r.Name, OwnerID: r.UserId, NodeID: r.NodeId, Status: r.Status, Suspended: r.Suspended, Memory: r.Memory, Disk: r.Disk, CPU: r.Cpu, PackageID: r.PackageId, PrimaryAllocation: r.PrimaryAllocation}, nil
}

func (a *API) ListServers() []*Server {
	r, _ := a.panel.ListServers(a.ctx(), &pb.ListServersRequest{})
	out := make([]*Server, len(r.GetServers()))
	for i, s := range r.GetServers() {
		out[i] = &Server{ID: s.Id, Name: s.Name, OwnerID: s.UserId, NodeID: s.NodeId, Status: s.Status, Suspended: s.Suspended, Memory: s.Memory, Disk: s.Disk, CPU: s.Cpu, PackageID: s.PackageId, PrimaryAllocation: s.PrimaryAllocation}
	}
	return out
}

func (a *API) ListServersByUser(userID string) []*Server {
	r, _ := a.panel.ListServers(a.ctx(), &pb.ListServersRequest{UserId: userID})
	out := make([]*Server, len(r.GetServers()))
	for i, s := range r.GetServers() {
		out[i] = &Server{ID: s.Id, Name: s.Name, OwnerID: s.UserId, NodeID: s.NodeId, Status: s.Status, Suspended: s.Suspended, Memory: s.Memory, Disk: s.Disk, CPU: s.Cpu, PackageID: s.PackageId, PrimaryAllocation: s.PrimaryAllocation}
	}
	return out
}

func (a *API) CreateServer(name, userID, nodeID, packageID string, memory, cpu, disk int32) (*Server, error) {
	r, err := a.panel.CreateServer(a.ctx(), &pb.CreateServerRequest{Name: name, UserId: userID, NodeId: nodeID, PackageId: packageID, Memory: memory, Cpu: cpu, Disk: disk})
	if err != nil {
		return nil, err
	}
	return &Server{ID: r.Id, Name: r.Name, OwnerID: r.UserId, NodeID: r.NodeId}, nil
}

func (a *API) UpdateServer(id string, name *string, memory, cpu, disk *int32) (*Server, error) {
	req := &pb.UpdateServerRequest{Id: id}
	if name != nil {
		req.Name = *name
	}
	if memory != nil {
		req.Memory = *memory
	}
	if cpu != nil {
		req.Cpu = *cpu
	}
	if disk != nil {
		req.Disk = *disk
	}
	r, err := a.panel.UpdateServer(a.ctx(), req)
	if err != nil {
		return nil, err
	}
	return &Server{ID: r.Id, Name: r.Name, OwnerID: r.UserId, NodeID: r.NodeId, Status: r.Status, Suspended: r.Suspended, Memory: r.Memory, Disk: r.Disk, CPU: r.Cpu, PackageID: r.PackageId, PrimaryAllocation: r.PrimaryAllocation}, nil
}

func (a *API) DeleteServer(id string) error {
	_, err := a.panel.DeleteServer(a.ctx(), &pb.IDRequest{Id: id})
	return err
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

func (a *API) ReinstallServer(id string) error {
	_, err := a.panel.ReinstallServer(a.ctx(), &pb.IDRequest{Id: id})
	return err
}

func (a *API) TransferServer(id, targetNodeID string) error {
	_, err := a.panel.TransferServer(a.ctx(), &pb.TransferServerRequest{ServerId: id, TargetNodeId: targetNodeID})
	return err
}

func (a *API) GetConsoleLog(serverID string, lines int32) ([]string, error) {
	r, err := a.panel.GetConsoleLog(a.ctx(), &pb.ConsoleLogRequest{ServerId: serverID, Lines: lines})
	if err != nil {
		return nil, err
	}
	return r.Lines, nil
}

func (a *API) SendCommand(serverID, command string) error {
	_, err := a.panel.SendCommand(a.ctx(), &pb.SendCommandRequest{ServerId: serverID, Command: command})
	return err
}

type ServerStats struct {
	MemoryBytes int64
	MemoryLimit int64
	CPUPercent  float64
	DiskBytes   int64
	NetworkRx   int64
	NetworkTx   int64
	State       string
}

func (a *API) GetServerStats(serverID string) (*ServerStats, error) {
	r, err := a.panel.GetServerStats(a.ctx(), &pb.IDRequest{Id: serverID})
	if err != nil {
		return nil, err
	}
	return &ServerStats{MemoryBytes: r.MemoryBytes, MemoryLimit: r.MemoryLimit, CPUPercent: r.CpuPercent, DiskBytes: r.DiskBytes, NetworkRx: r.NetworkRx, NetworkTx: r.NetworkTx, State: r.State}, nil
}

func (a *API) AddAllocation(serverID string, port int32) error {
	_, err := a.panel.AddAllocation(a.ctx(), &pb.AllocationRequest{ServerId: serverID, Port: port})
	return err
}

func (a *API) DeleteAllocation(serverID string, port int32) error {
	_, err := a.panel.DeleteAllocation(a.ctx(), &pb.AllocationRequest{ServerId: serverID, Port: port})
	return err
}

func (a *API) SetPrimaryAllocation(serverID string, port int32) error {
	_, err := a.panel.SetPrimaryAllocation(a.ctx(), &pb.AllocationRequest{ServerId: serverID, Port: port})
	return err
}

func (a *API) UpdateServerVariables(serverID string, variables map[string]string) error {
	_, err := a.panel.UpdateServerVariables(a.ctx(), &pb.UpdateVariablesRequest{ServerId: serverID, Variables: variables})
	return err
}

func (a *API) CompressFiles(serverID string, paths []string, destination string) error {
	_, err := a.panel.CompressFiles(a.ctx(), &pb.CompressRequest{ServerId: serverID, Paths: paths, Destination: destination})
	return err
}

func (a *API) DecompressFile(serverID, path string) error {
	_, err := a.panel.DecompressFile(a.ctx(), &pb.FilePathRequest{ServerId: serverID, Path: path})
	return err
}

func (a *API) GetUser(id string) (*User, error) {
	r, err := a.panel.GetUser(a.ctx(), &pb.IDRequest{Id: id})
	if err != nil {
		return nil, err
	}
	return &User{ID: r.Id, Username: r.Username, Email: r.Email, IsAdmin: r.IsAdmin, IsBanned: r.IsBanned, ForcePasswordReset: r.ForcePasswordReset, RamLimit: r.RamLimit, CpuLimit: r.CpuLimit, DiskLimit: r.DiskLimit, ServerLimit: r.ServerLimit, CreatedAt: r.CreatedAt}, nil
}

func (a *API) GetUserByEmail(email string) (*User, error) {
	r, err := a.panel.GetUserByEmail(a.ctx(), &pb.EmailRequest{Email: email})
	if err != nil {
		return nil, err
	}
	return &User{ID: r.Id, Username: r.Username, Email: r.Email, IsAdmin: r.IsAdmin, IsBanned: r.IsBanned, ForcePasswordReset: r.ForcePasswordReset, RamLimit: r.RamLimit, CpuLimit: r.CpuLimit, DiskLimit: r.DiskLimit, ServerLimit: r.ServerLimit, CreatedAt: r.CreatedAt}, nil
}

func (a *API) GetUserByUsername(username string) (*User, error) {
	r, err := a.panel.GetUserByUsername(a.ctx(), &pb.UsernameRequest{Username: username})
	if err != nil {
		return nil, err
	}
	return &User{ID: r.Id, Username: r.Username, Email: r.Email, IsAdmin: r.IsAdmin, IsBanned: r.IsBanned, ForcePasswordReset: r.ForcePasswordReset, RamLimit: r.RamLimit, CpuLimit: r.CpuLimit, DiskLimit: r.DiskLimit, ServerLimit: r.ServerLimit, CreatedAt: r.CreatedAt}, nil
}

func (a *API) ListUsers() []*User {
	r, _ := a.panel.ListUsers(a.ctx(), &pb.ListUsersRequest{})
	out := make([]*User, len(r.GetUsers()))
	for i, u := range r.GetUsers() {
		out[i] = &User{ID: u.Id, Username: u.Username, Email: u.Email, IsAdmin: u.IsAdmin, IsBanned: u.IsBanned, ForcePasswordReset: u.ForcePasswordReset, RamLimit: u.RamLimit, CpuLimit: u.CpuLimit, DiskLimit: u.DiskLimit, ServerLimit: u.ServerLimit, CreatedAt: u.CreatedAt}
	}
	return out
}

func (a *API) CreateUser(email, username, password string) (*User, error) {
	r, err := a.panel.CreateUser(a.ctx(), &pb.CreateUserRequest{Email: email, Username: username, Password: password})
	if err != nil {
		return nil, err
	}
	return &User{ID: r.Id, Username: r.Username, Email: r.Email}, nil
}

func (a *API) UpdateUser(id string, username, email *string) (*User, error) {
	req := &pb.UpdateUserRequest{Id: id}
	if username != nil {
		req.Username = *username
	}
	if email != nil {
		req.Email = *email
	}
	r, err := a.panel.UpdateUser(a.ctx(), req)
	if err != nil {
		return nil, err
	}
	return &User{ID: r.Id, Username: r.Username, Email: r.Email, IsAdmin: r.IsAdmin, IsBanned: r.IsBanned, ForcePasswordReset: r.ForcePasswordReset, RamLimit: r.RamLimit, CpuLimit: r.CpuLimit, DiskLimit: r.DiskLimit, ServerLimit: r.ServerLimit, CreatedAt: r.CreatedAt}, nil
}

func (a *API) DeleteUser(id string) error {
	_, err := a.panel.DeleteUser(a.ctx(), &pb.IDRequest{Id: id})
	return err
}

func (a *API) BanUser(id string) error {
	_, err := a.panel.BanUser(a.ctx(), &pb.IDRequest{Id: id})
	return err
}

func (a *API) UnbanUser(id string) error {
	_, err := a.panel.UnbanUser(a.ctx(), &pb.IDRequest{Id: id})
	return err
}

func (a *API) SetAdmin(id string) error {
	_, err := a.panel.SetAdmin(a.ctx(), &pb.IDRequest{Id: id})
	return err
}

func (a *API) RevokeAdmin(id string) error {
	_, err := a.panel.RevokeAdmin(a.ctx(), &pb.IDRequest{Id: id})
	return err
}

func (a *API) ForcePasswordReset(id string) error {
	_, err := a.panel.ForcePasswordReset(a.ctx(), &pb.IDRequest{Id: id})
	return err
}

func (a *API) SetUserResources(id string, ram, cpu, disk, servers *int32) error {
	req := &pb.SetUserResourcesRequest{UserId: id}
	if ram != nil {
		req.RamLimit = *ram
	}
	if cpu != nil {
		req.CpuLimit = *cpu
	}
	if disk != nil {
		req.DiskLimit = *disk
	}
	if servers != nil {
		req.ServerLimit = *servers
	}
	_, err := a.panel.SetUserResources(a.ctx(), req)
	return err
}

func (a *API) ListNodes() []*Node {
	r, _ := a.panel.ListNodes(a.ctx(), &pb.Empty{})
	out := make([]*Node, len(r.GetNodes()))
	for i, n := range r.GetNodes() {
		out[i] = &Node{ID: n.Id, Name: n.Name, FQDN: n.Fqdn, Port: n.Port, IsOnline: n.IsOnline, LastHeartbeat: n.LastHeartbeat}
	}
	return out
}

func (a *API) GetNode(id string) (*Node, error) {
	r, err := a.panel.GetNode(a.ctx(), &pb.IDRequest{Id: id})
	if err != nil {
		return nil, err
	}
	return &Node{ID: r.Id, Name: r.Name, FQDN: r.Fqdn, Port: r.Port, IsOnline: r.IsOnline, LastHeartbeat: r.LastHeartbeat}, nil
}

func (a *API) CreateNode(name, fqdn string, port int32) (*Node, string, error) {
	r, err := a.panel.CreateNode(a.ctx(), &pb.CreateNodeRequest{Name: name, Fqdn: fqdn, Port: port})
	if err != nil {
		return nil, "", err
	}
	return &Node{ID: r.Node.Id, Name: r.Node.Name, FQDN: r.Node.Fqdn, Port: r.Node.Port, LastHeartbeat: r.Node.LastHeartbeat}, r.Token, nil
}

func (a *API) DeleteNode(id string) error {
	_, err := a.panel.DeleteNode(a.ctx(), &pb.IDRequest{Id: id})
	return err
}

func (a *API) ResetNodeToken(id string) (string, error) {
	r, err := a.panel.ResetNodeToken(a.ctx(), &pb.IDRequest{Id: id})
	if err != nil {
		return "", err
	}
	return r.Token, nil
}

func (a *API) ListFiles(serverID, path string) []*File {
	r, _ := a.panel.ListFiles(a.ctx(), &pb.FilePathRequest{ServerId: serverID, Path: path})
	out := make([]*File, len(r.GetFiles()))
	for i, f := range r.GetFiles() {
		out[i] = &File{Name: f.Name, Size: f.Size, IsDir: f.IsDir, ModTime: f.Modified, Mime: f.Mime}
	}
	return out
}

func (a *API) ReadFile(serverID, path string) ([]byte, error) {
	r, err := a.panel.ReadFile(a.ctx(), &pb.FilePathRequest{ServerId: serverID, Path: path})
	if err != nil {
		return nil, err
	}
	return r.Content, nil
}

func (a *API) WriteFile(serverID, path string, content []byte) error {
	_, err := a.panel.WriteFile(a.ctx(), &pb.WriteFileRequest{ServerId: serverID, Path: path, Content: content})
	return err
}

func (a *API) DeleteFile(serverID, path string) error {
	_, err := a.panel.DeleteFile(a.ctx(), &pb.FilePathRequest{ServerId: serverID, Path: path})
	return err
}

func (a *API) CreateFolder(serverID, path string) error {
	_, err := a.panel.CreateFolder(a.ctx(), &pb.FilePathRequest{ServerId: serverID, Path: path})
	return err
}

func (a *API) MoveFile(serverID, from, to string) error {
	_, err := a.panel.MoveFile(a.ctx(), &pb.MoveFileRequest{ServerId: serverID, From: from, To: to})
	return err
}

func (a *API) CopyFile(serverID, from, to string) error {
	_, err := a.panel.CopyFile(a.ctx(), &pb.MoveFileRequest{ServerId: serverID, From: from, To: to})
	return err
}

func (a *API) ListDatabases(serverID string) []*Database {
	r, _ := a.panel.ListDatabases(a.ctx(), &pb.IDRequest{Id: serverID})
	out := make([]*Database, len(r.GetDatabases()))
	for i, d := range r.GetDatabases() {
		out[i] = &Database{ID: d.Id, Name: d.Name, Username: d.Username, Host: d.Host, Port: d.Port}
	}
	return out
}

func (a *API) CreateDatabase(serverID, name string) (*Database, error) {
	r, err := a.panel.CreateDatabase(a.ctx(), &pb.CreateDatabaseRequest{ServerId: serverID, Name: name})
	if err != nil {
		return nil, err
	}
	return &Database{ID: r.Id, Name: r.Name, Username: r.Username, Host: r.Host, Port: r.Port, Password: r.Password}, nil
}

func (a *API) DeleteDatabase(id string) error {
	_, err := a.panel.DeleteDatabase(a.ctx(), &pb.IDRequest{Id: id})
	return err
}

func (a *API) RotateDatabasePassword(id string) (*Database, error) {
	r, err := a.panel.RotateDatabasePassword(a.ctx(), &pb.IDRequest{Id: id})
	if err != nil {
		return nil, err
	}
	return &Database{ID: r.Id, Name: r.Name, Username: r.Username, Host: r.Host, Port: r.Port, Password: r.Password}, nil
}

func (a *API) ListDatabaseHosts() []*DatabaseHost {
	r, _ := a.panel.ListDatabaseHosts(a.ctx(), &pb.Empty{})
	out := make([]*DatabaseHost, len(r.GetHosts()))
	for i, h := range r.GetHosts() {
		out[i] = &DatabaseHost{ID: h.Id, Name: h.Name, Host: h.Host, Port: h.Port, Username: h.Username, MaxDatabases: h.MaxDatabases, DatabasesCount: h.DatabasesCount}
	}
	return out
}

func (a *API) ListBackups(serverID string) []*Backup {
	r, _ := a.panel.ListBackups(a.ctx(), &pb.IDRequest{Id: serverID})
	out := make([]*Backup, len(r.GetBackups()))
	for i, b := range r.GetBackups() {
		out[i] = &Backup{ID: b.Id, Name: b.Name, Size: b.Size, CreatedAt: b.CreatedAt}
	}
	return out
}

func (a *API) CreateBackup(serverID, name string) error {
	_, err := a.panel.CreateBackup(a.ctx(), &pb.CreateBackupRequest{ServerId: serverID, Name: name})
	return err
}

func (a *API) DeleteBackup(serverID, backupID string) error {
	_, err := a.panel.DeleteBackup(a.ctx(), &pb.DeleteBackupRequest{ServerId: serverID, BackupId: backupID})
	return err
}

func (a *API) ListPackages() []*Package {
	r, _ := a.panel.ListPackages(a.ctx(), &pb.Empty{})
	out := make([]*Package, len(r.GetPackages()))
	for i, p := range r.GetPackages() {
		out[i] = &Package{ID: p.Id, Name: p.Name, Description: p.Description, DockerImage: p.DockerImage, StartupCommand: p.StartupCommand, StopCommand: p.StopCommand, ConfigFiles: p.ConfigFiles, Memory: p.DefaultMemory, CPU: p.DefaultCpu, Disk: p.DefaultDisk, IsPublic: p.IsPublic}
	}
	return out
}

func (a *API) GetPackage(id string) (*Package, error) {
	r, err := a.panel.GetPackage(a.ctx(), &pb.IDRequest{Id: id})
	if err != nil {
		return nil, err
	}
	return &Package{ID: r.Id, Name: r.Name, Description: r.Description, DockerImage: r.DockerImage, StartupCommand: r.StartupCommand, StopCommand: r.StopCommand, ConfigFiles: r.ConfigFiles, Memory: r.DefaultMemory, CPU: r.DefaultCpu, Disk: r.DefaultDisk, IsPublic: r.IsPublic}, nil
}

func (a *API) CreatePackage(name, description, dockerImage, startupCmd, stopCmd, configFiles string, memory, cpu, disk int32, isPublic bool) (*Package, error) {
	r, err := a.panel.CreatePackage(a.ctx(), &pb.CreatePackageRequest{
		Name: name, Description: description, DockerImage: dockerImage,
		StartupCommand: startupCmd, StopCommand: stopCmd, ConfigFiles: configFiles,
		DefaultMemory: memory, DefaultCpu: cpu, DefaultDisk: disk, IsPublic: isPublic,
	})
	if err != nil {
		return nil, err
	}
	return &Package{ID: r.Id, Name: r.Name, Description: r.Description, DockerImage: r.DockerImage, StartupCommand: r.StartupCommand, StopCommand: r.StopCommand, ConfigFiles: r.ConfigFiles, Memory: r.DefaultMemory, CPU: r.DefaultCpu, Disk: r.DefaultDisk, IsPublic: r.IsPublic}, nil
}

func (a *API) UpdatePackage(id string, name, description *string, memory, cpu, disk *int32) (*Package, error) {
	req := &pb.UpdatePackageRequest{Id: id}
	if name != nil {
		req.Name = *name
	}
	if description != nil {
		req.Description = *description
	}
	if memory != nil {
		req.DefaultMemory = *memory
	}
	if cpu != nil {
		req.DefaultCpu = *cpu
	}
	if disk != nil {
		req.DefaultDisk = *disk
	}
	r, err := a.panel.UpdatePackage(a.ctx(), req)
	if err != nil {
		return nil, err
	}
	return &Package{ID: r.Id, Name: r.Name, Description: r.Description, DockerImage: r.DockerImage, StartupCommand: r.StartupCommand, StopCommand: r.StopCommand, ConfigFiles: r.ConfigFiles, Memory: r.DefaultMemory, CPU: r.DefaultCpu, Disk: r.DefaultDisk, IsPublic: r.IsPublic}, nil
}

func (a *API) DeletePackage(id string) error {
	_, err := a.panel.DeletePackage(a.ctx(), &pb.IDRequest{Id: id})
	return err
}

func (a *API) ListIPBans() []*IPBan {
	r, _ := a.panel.ListIPBans(a.ctx(), &pb.Empty{})
	out := make([]*IPBan, len(r.GetBans()))
	for i, b := range r.GetBans() {
		out[i] = &IPBan{ID: b.Id, IP: b.Ip, Reason: b.Reason, CreatedAt: b.CreatedAt}
	}
	return out
}

func (a *API) CreateIPBan(ip, reason string) (*IPBan, error) {
	r, err := a.panel.CreateIPBan(a.ctx(), &pb.CreateIPBanRequest{Ip: ip, Reason: reason})
	if err != nil {
		return nil, err
	}
	return &IPBan{ID: r.Id, IP: r.Ip, Reason: r.Reason, CreatedAt: r.CreatedAt}, nil
}

func (a *API) DeleteIPBan(id string) error {
	_, err := a.panel.DeleteIPBan(a.ctx(), &pb.IDRequest{Id: id})
	return err
}

func (a *API) ListSubusers(serverID string) []*Subuser {
	r, _ := a.panel.ListSubusers(a.ctx(), &pb.IDRequest{Id: serverID})
	out := make([]*Subuser, len(r.GetSubusers()))
	for i, s := range r.GetSubusers() {
		out[i] = &Subuser{ID: s.Id, UserID: s.UserId, Username: s.Username, Email: s.Email, Permissions: s.Permissions}
	}
	return out
}

func (a *API) AddSubuser(serverID, email string, permissions []string) (*Subuser, error) {
	r, err := a.panel.AddSubuser(a.ctx(), &pb.AddSubuserRequest{ServerId: serverID, Email: email, Permissions: permissions})
	if err != nil {
		return nil, err
	}
	return &Subuser{ID: r.Id, UserID: r.UserId, Username: r.Username, Email: r.Email, Permissions: r.Permissions}, nil
}

func (a *API) UpdateSubuser(serverID, subuserID string, permissions []string) error {
	_, err := a.panel.UpdateSubuser(a.ctx(), &pb.UpdateSubuserRequest{ServerId: serverID, SubuserId: subuserID, Permissions: permissions})
	return err
}

func (a *API) RemoveSubuser(serverID, subuserID string) error {
	_, err := a.panel.RemoveSubuser(a.ctx(), &pb.RemoveSubuserRequest{ServerId: serverID, SubuserId: subuserID})
	return err
}

func (a *API) GetSettings() *Settings {
	r, _ := a.panel.GetSettings(a.ctx(), &pb.Empty{})
	return &Settings{RegistrationEnabled: r.GetRegistrationEnabled(), ServerCreationEnabled: r.GetServerCreationEnabled()}
}

func (a *API) SetRegistrationEnabled(enabled bool) error {
	_, err := a.panel.SetRegistrationEnabled(a.ctx(), &pb.BoolRequest{Value: enabled})
	return err
}

func (a *API) SetServerCreationEnabled(enabled bool) error {
	_, err := a.panel.SetServerCreationEnabled(a.ctx(), &pb.BoolRequest{Value: enabled})
	return err
}

func (a *API) GetActivityLogs(limit int32) []*ActivityLog {
	r, _ := a.panel.GetActivityLogs(a.ctx(), &pb.GetLogsRequest{Limit: limit})
	out := make([]*ActivityLog, len(r.GetLogs()))
	for i, l := range r.GetLogs() {
		out[i] = &ActivityLog{ID: l.Id, UserID: l.UserId, Username: l.Username, Action: l.Action, Description: l.Description, IP: l.Ip, IsAdmin: l.IsAdmin, CreatedAt: l.CreatedAt}
	}
	return out
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

func (a *API) QueryDB(query string, args ...string) ([]map[string]interface{}, error) {
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
}

func (a *API) BroadcastEvent(eventType string, data map[string]string) {
	a.panel.BroadcastEvent(a.ctx(), &pb.BroadcastEventRequest{EventType: eventType, Data: data})
}

type HTTPResponse struct {
	Status  int
	Headers map[string]string
	Body    []byte
	Error   string
}

func (a *API) HTTP(method, url string, headers map[string]string, body []byte) *HTTPResponse {
	r, err := a.panel.HTTPRequest(a.ctx(), &pb.PluginHTTPRequest{
		Method:  method,
		Url:     url,
		Headers: headers,
		Body:    body,
	})
	if err != nil {
		return &HTTPResponse{Error: err.Error()}
	}
	return &HTTPResponse{Status: int(r.Status), Headers: r.Headers, Body: r.Body, Error: r.Error}
}

func (a *API) HTTPGet(url string, headers map[string]string) *HTTPResponse {
	return a.HTTP("GET", url, headers, nil)
}

func (a *API) HTTPPost(url string, headers map[string]string, body []byte) *HTTPResponse {
	return a.HTTP("POST", url, headers, body)
}

func (a *API) HTTPPut(url string, headers map[string]string, body []byte) *HTTPResponse {
	return a.HTTP("PUT", url, headers, body)
}

func (a *API) HTTPDelete(url string, headers map[string]string) *HTTPResponse {
	return a.HTTP("DELETE", url, headers, nil)
}

func (a *API) CallPlugin(pluginID, method string, data []byte) ([]byte, error) {
	r, err := a.panel.CallPlugin(a.ctx(), &pb.CallPluginRequest{
		PluginId: pluginID,
		Method:   method,
		Data:     data,
	})
	if err != nil {
		return nil, err
	}
	if r.Error != "" {
		return nil, fmt.Errorf(r.Error)
	}
	return r.Data, nil
}
