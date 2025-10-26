package object

import (
	"strings"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	psutilnet "github.com/shirou/gopsutil/v3/net"
)

var PsutilBuiltins = NewBuiltinSliceType{
	{Name: "_cpu_usage_percent", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 0 {
				return newInvalidArgCountError("cpu_usage_percent", len(args), 0, "")
			}
			usages, err := cpu.Percent(0, true)
			if err != nil {
				return newError("`cpu_usage_percent` error: %s", err.Error())
			}
			l := &List{Elements: make([]Object, len(usages))}
			for i, v := range usages {
				l.Elements[i] = &Float{Value: v}
			}
			return l
		},
		HelpStr: helpStrArgs{
			explanation: "`cpu_usage_percent` returns a list of cpu usages as floats per core",
			signature:   "cpu_usage_percent() -> list[float]",
			errors:      "InvalidArgCount,CustomError",
			example:     "cpu_usage_percent() => [1.0,0.4,0.2,0.6]",
		}.String(),
	}},
	{Name: "_cpu_info", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 0 {
				return newInvalidArgCountError("cpu_info", len(args), 0, "")
			}
			infos, err := cpu.Info()
			if err != nil {
				return newError("`cpu_info` error: %s", err.Error())
			}
			l := &List{Elements: make([]Object, len(infos))}
			for i, v := range infos {
				l.Elements[i] = &Stringo{Value: v.String()}
			}
			return l
		},
		HelpStr: helpStrArgs{
			explanation: "`cpu_info` returns a list of json strings of cpu info per prcoessor",
			signature:   "cpu_info() -> list[str]",
			errors:      "InvalidArgCount,CustomError",
			example:     "cpu_info() => [json_with_keys('cpu','vendorId','family','model','stepping','physicalId','coreId','cores','modelName','mhz','cacheSize','flags','microcode')]",
		}.String(),
	}},
	{Name: "_cpu_time_info", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 0 {
				return newInvalidArgCountError("cpu_time_info", len(args), 0, "")
			}
			infos, err := cpu.Times(true)
			if err != nil {
				return newError("`cpu_time_info` error: %s", err.Error())
			}
			l := &List{Elements: make([]Object, len(infos))}
			for i, v := range infos {
				l.Elements[i] = &Stringo{Value: v.String()}
			}
			return l
		},
		HelpStr: helpStrArgs{
			explanation: "`cpu_time_info` returns a list of json strings of cpu time stat info per prcoessor",
			signature:   "cpu_time_info() -> list[str]",
			errors:      "InvalidArgCount,CustomError",
			example:     "cpu_time_info() => [json_with_keys('cpu','user','system','idle','nice','iowait','irq','softirq','steal','guest','guestNice')]",
		}.String(),
	}},
	{Name: "_cpu_count", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 0 {
				return newInvalidArgCountError("cpu_count", len(args), 0, "")
			}
			count, err := cpu.Counts(true)
			if err != nil {
				return newError("`cpu_count` error: %s", err.Error())
			}
			return &Integer{Value: int64(count)}
		},
		HelpStr: helpStrArgs{
			explanation: "`cpu_count` returns the number of cores as an INTEGER",
			signature:   "cpu_count() -> int",
			errors:      "InvalidArgCount,CustomError",
			example:     "cpu_count() => 4",
		}.String(),
	}},
	{Name: "_mem_virt_info", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 0 {
				return newInvalidArgCountError("mem_virt_info", len(args), 0, "")
			}
			v, err := mem.VirtualMemory()
			if err != nil {
				return newError("`mem_virt_info` error: %s", err.Error())
			}
			return &Stringo{Value: v.String()}
		},
		HelpStr: helpStrArgs{
			explanation: "`mem_virt_info` returns a json string of virtual memory info",
			signature:   "mem_virt_info() -> str",
			errors:      "InvalidArgCount,CustomError",
			example:     "mem_virt_info() => json_with_keys('total','available','used','usedPercent','free','active','inactive','wired','laundry','buffers','cached','writeBack','dirty','writeBackTmp','shared','slab','sreclaimable','sunreclaim','pageTables','swapCached','commitLimit','committedAS','highTotal','highFree','lowTotal','lowFree','swapTotal','swapFree','mapped','vmallocTotal','vmallocUsed','vmallocChunk','hugePagesTotal','hugePagesFree','hugePagesRsvd','hugePagesSurp','hugePageSize')",
		}.String(),
	}},
	{Name: "_mem_swap_info", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 0 {
				return newInvalidArgCountError("mem_swap_info", len(args), 0, "")
			}
			v, err := mem.SwapMemory()
			if err != nil {
				return newError("`mem_swap_info` error: %s", err.Error())
			}
			return &Stringo{Value: v.String()}
		},
		HelpStr: helpStrArgs{
			explanation: "`mem_swap_info` returns a json string of swap memory info",
			signature:   "mem_swap_info() -> str",
			errors:      "InvalidArgCount,CustomError",
			example:     "mem_swap_info() => json_with_keys('total','used','free','usedPercent','sin','sout','pgIn','pgOut','pgFault','pgMajFault')",
		}.String(),
	}},
	{Name: "_host_info", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 0 {
				return newInvalidArgCountError("host_info", len(args), 0, "")
			}
			i, err := host.Info()
			if err != nil {
				return newError("`host_info` error: %s", err.Error())
			}
			return &Stringo{Value: i.String()}
		},
		HelpStr: helpStrArgs{
			explanation: "`host_info` returns a json string of host info",
			signature:   "host_info() -> str",
			errors:      "InvalidArgCount,CustomError",
			example:     "host_info() => json_with_keys('hostname','uptime','bootTime','procs','os','platform','platformFamily','platformVersion','kernelVersion','kernelArch','virtualizationSystem','virtualizationRole','hostId')",
		}.String(),
	}},
	{Name: "_host_temps_info", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 0 {
				return newInvalidArgCountError("host_temps_info", len(args), 0, "")
			}
			temps, err := host.SensorsTemperatures()
			if err != nil {
				if !strings.Contains(err.Error(), "warnings") {
					return newError("`host_temps_info` error: %s", err.Error())
				}
			}
			l := &List{Elements: make([]Object, len(temps))}
			for i, t := range temps {
				l.Elements[i] = &Stringo{Value: t.String()}
			}
			return l
		},
		HelpStr: helpStrArgs{
			explanation: "`host_temps_info` returns a list of json strings of host sensor temperature info",
			signature:   "host_temps_info() -> list[str]",
			errors:      "InvalidArgCount,CustomError",
			example:     "host_temps_info() => [json_with_keys('sensorKey','temperature','sensorHigh','sensorCritical')]",
		}.String(),
	}},
	{Name: "_net_connections", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("net_connections", len(args), 1, "")
			}
			if args[0].Type() != STRING_OBJ {
				return newPositionalTypeError("net_connections", 1, STRING_OBJ, args[0].Type())
			}
			option := args[0].(*Stringo).Value
			conns, err := psutilnet.Connections(option)
			if err != nil {
				return newError("`net_connections` error: %s", err.Error())
			}
			l := &List{Elements: make([]Object, len(conns))}
			for i, c := range conns {
				l.Elements[i] = &Stringo{Value: c.String()}
			}
			return l
		},
		HelpStr: helpStrArgs{
			explanation: "`net_connections` returns a list of json strings of host network connection stats for the given option",
			signature:   "net_connections(option: str('all'|'tcp'|'tcp4'|'tcp6'|'udp'|'udp4'|'udp6'|'inet'|'inet4'|'inet6')='all') -> list[str]",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "net_connections() => [json_with_keys('fd','family','type','localaddr','remoteaddr','status','uids','pid')]",
		}.String(),
	}},
	{Name: "_net_io_info", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 0 {
				return newInvalidArgCountError("net_io_info", len(args), 0, "")
			}
			ioc, err := psutilnet.IOCounters(true)
			if err != nil {
				return newError("`net_io_info` error: %s", err.Error())
			}
			l := &List{Elements: make([]Object, len(ioc))}
			for i, oc := range ioc {
				l.Elements[i] = &Stringo{Value: oc.String()}
			}
			return l
		},
		HelpStr: helpStrArgs{
			explanation: "`net_io_info` returns a list of json strings of network io stat info",
			signature:   "net_io_info() -> list[str]",
			errors:      "InvalidArgCount,CustomError",
			example:     "net_io_info() => [json_with_keys('name','bytesSent','bytesRecv','packetsSent','packetsRecv','errin','errout','dropin','dropout','fifoin','fifoout')]",
		}.String(),
	}},
	{Name: "_disk_partitions", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 0 {
				return newInvalidArgCountError("disk_partitions", len(args), 0, "")
			}
			parts, err := disk.Partitions(true)
			if err != nil {
				return newError("`disk_partitions` error: %s", err.Error())
			}
			l := &List{Elements: make([]Object, len(parts))}
			for i, p := range parts {
				l.Elements[i] = &Stringo{Value: p.String()}
			}
			return l
		},
		HelpStr: helpStrArgs{
			explanation: "`disk_partitions` returns a list of json strings of disk partition info",
			signature:   "disk_partitions() -> list[str]",
			errors:      "InvalidArgCount,CustomError",
			example:     "disk_partitions() => [json_with_keys('device','mountpoint','fstype','opts')]",
		}.String(),
	}},
	{Name: "_disk_io_info", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 0 {
				return newInvalidArgCountError("disk_io_info", len(args), 0, "")
			}
			ioc, err := disk.IOCounters()
			if err != nil {
				return newError("`disk_io_info` error: %s", err.Error())
			}
			m := NewOrderedMap[string, Object]()
			for k, v := range ioc {
				m.Set(k, &Stringo{Value: v.String()})
			}
			return CreateMapObjectForGoMap(*m)
		},
		HelpStr: helpStrArgs{
			explanation: "`disk_io_info` returns a map of drive to json string of disk io info",
			signature:   "disk_io_info() -> map[str:str]",
			errors:      "InvalidArgCount,CustomError",
			example:     "disk_io_info() => {'drive': json_with_keys('readCount','mergedReadCount','writeCount','mergedWriteCount','readBytes','writeBytes','readTime','writeTime','iopsInProgress','ioTime','weightedIO','name','serialNumber','label')...}",
		}.String(),
	}},
	{Name: "_disk_usage", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("disk_usage", len(args), 1, "")
			}
			if args[0].Type() != STRING_OBJ {
				return newPositionalTypeError("disk_usage", 1, STRING_OBJ, args[0].Type())
			}
			path := args[0].(*Stringo).Value
			usage, err := disk.Usage(path)
			if err != nil {
				return newError("`disk_usage` error: %s", err.Error())
			}
			return &Stringo{Value: usage.String()}
		},
		HelpStr: helpStrArgs{
			explanation: "`disk_usage` returns a json string of disk usage for the given path",
			signature:   "disk_usage(path: str) -> str",
			errors:      "InvalidArgCount,CustomError",
			example:     "disk_usage(root_path) => json_with_keys('path','fstype','total','free','used','usedPercent','inodesTotal','inodesUsed','inodesFree','inodesUsedPercent')",
		}.String(),
	}},
}
