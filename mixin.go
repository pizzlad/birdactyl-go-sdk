package birdactyl

import "encoding/json"

const (
	MixinServerCreate    = "server.create"
	MixinServerUpdate    = "server.update"
	MixinServerDelete    = "server.delete"
	MixinServerStart     = "server.start"
	MixinServerStop      = "server.stop"
	MixinServerRestart   = "server.restart"
	MixinServerKill      = "server.kill"
	MixinServerSuspend   = "server.suspend"
	MixinServerUnsuspend = "server.unsuspend"
	MixinServerReinstall = "server.reinstall"
	MixinServerTransfer  = "server.transfer"

	MixinUserCreate       = "user.create"
	MixinUserUpdate       = "user.update"
	MixinUserDelete       = "user.delete"
	MixinUserAuthenticate = "user.authenticate"
	MixinUserBan          = "user.ban"
	MixinUserUnban        = "user.unban"

	MixinDatabaseCreate = "database.create"
	MixinDatabaseDelete = "database.delete"

	MixinBackupCreate = "backup.create"
	MixinBackupDelete = "backup.delete"

	MixinFileRead       = "file.read"
	MixinFileWrite      = "file.write"
	MixinFileDelete     = "file.delete"
	MixinFileUpload     = "file.upload"
	MixinFileMove       = "file.move"
	MixinFileCopy       = "file.copy"
	MixinFileCompress   = "file.compress"
	MixinFileDecompress = "file.decompress"

	MixinNodeCreate = "node.create"
	MixinNodeDelete = "node.delete"

	MixinPackageCreate = "package.create"
	MixinPackageUpdate = "package.update"
	MixinPackageDelete = "package.delete"

	MixinSubuserAdd    = "subuser.add"
	MixinSubuserUpdate = "subuser.update"
	MixinSubuserRemove = "subuser.remove"

	MixinIPBanCreate = "ipban.create"
	MixinIPBanDelete = "ipban.delete"

	MixinAllocationAdd        = "allocation.add"
	MixinAllocationDelete     = "allocation.delete"
	MixinAllocationSetPrimary = "allocation.set_primary"

	MixinDBHostCreate = "dbhost.create"
	MixinDBHostUpdate = "dbhost.update"
	MixinDBHostDelete = "dbhost.delete"

	MixinSettingsUpdate = "settings.update"
)

type MixinContext struct {
	Target    string
	RequestID string
	input     map[string]interface{}
	chainData map[string]interface{}
	nextCalled bool
	result     MixinResult
}

type MixinResult struct {
	action        int
	output        map[string]interface{}
	err           string
	modifiedInput map[string]interface{}
}

type MixinHandler func(*MixinContext) MixinResult

func (c *MixinContext) Get(key string) interface{} {
	return c.input[key]
}

func (c *MixinContext) GetString(key string) string {
	if v, ok := c.input[key].(string); ok {
		return v
	}
	return ""
}

func (c *MixinContext) GetInt(key string) int {
	switch v := c.input[key].(type) {
	case int:
		return v
	case float64:
		return int(v)
	case int64:
		return int(v)
	}
	return 0
}

func (c *MixinContext) GetBool(key string) bool {
	if v, ok := c.input[key].(bool); ok {
		return v
	}
	return false
}

func (c *MixinContext) Input() map[string]interface{} {
	return c.input
}

func (c *MixinContext) Set(key string, value interface{}) {
	if c.result.modifiedInput == nil {
		c.result.modifiedInput = make(map[string]interface{})
		for k, v := range c.input {
			c.result.modifiedInput[k] = v
		}
	}
	c.result.modifiedInput[key] = value
}

func (c *MixinContext) ChainData() map[string]interface{} {
	return c.chainData
}

func (c *MixinContext) Next() MixinResult {
	c.nextCalled = true
	return MixinResult{action: 0, modifiedInput: c.result.modifiedInput}
}

func (c *MixinContext) Return(data interface{}) MixinResult {
	var out map[string]interface{}
	switch v := data.(type) {
	case map[string]interface{}:
		out = v
	default:
		b, _ := json.Marshal(data)
		json.Unmarshal(b, &out)
	}
	return MixinResult{action: 1, output: out}
}

func (c *MixinContext) Error(msg string) MixinResult {
	return MixinResult{action: 2, err: msg}
}

type MixinRegistration struct {
	Target   string
	Priority int
	Handler  MixinHandler
}
