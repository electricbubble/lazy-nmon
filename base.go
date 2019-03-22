package lazynmon

// ZZZZ nmon的时间
type ZZZZ struct {
	TName string
	Time  string
	Date  string
}

// CPUAll 所有CPU概述，显示所有CPU平均占用情况，其中包含SMT状态
type CPUAll struct {
	TName     string
	UserUsage float64
	SysUsage  float64
	WaitUsage float64
	Usage     float64
	// IdleUsage string
	// BusyUsage string
	// CPUs      string
}

// DiskRead 每个hdisk的平均读情况
type DiskRead struct {
	TName     string
	ReadRatio float64
}

// DiskWrite 每个hdisk的平均写情况
type DiskWrite struct {
	TName      string
	WriteRatio float64
}

// Memory 内存使用情况
type Memory struct {
	TName   string
	Total   float64
	Free    float64
	Cached  float64
	Buffers float64
	Usage   float64
}

// Net 网络使用情况
type Net struct {
	TName      string
	ReadTotal  float64
	WriteTotal float64
}
