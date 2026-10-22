package windowsplatform

type GameUDPGuardStatus struct {
	DriverAntiCheat string   `json:"driver_anti_cheat,omitempty"`
	ForeignPIDs     []uint32 `json:"foreign_pids"`
	Ready           bool     `json:"ready"`
	Conflict        bool     `json:"conflict"`
	ForeignHandles  int      `json:"foreign_handles"`
	Error           string   `json:"error,omitempty"`
}
