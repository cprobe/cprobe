package tomcat

type ResponseStruct struct {
	Tomcat TomcatStruct `json:"tomcat"`
}

type JvmMemory struct {
	Free  string `json:"free"`
	Total string `json:"total"`
	Max   string `json:"max"`
}

type Memorypool struct {
	Name           string `json:"name"`
	Type           string `json:"type"`
	UsageInit      string `json:"usageInit"`
	UsageCommitted string `json:"usageCommitted"`
	UsageMax       string `json:"usageMax"`
	UsageUsed      string `json:"usageUsed"`
}

type TomcatJvm struct {
	JvmMemory      JvmMemory    `json:"memory"`
	JvmMemoryPools []Memorypool `json:"memorypool"`
}

type ThreadInfo struct {
	MaxThreads         string `json:"maxThreads"`
	CurrentThreadCount string `json:"currentThreadCount"`
	CurrentThreadsBusy string `json:"currentThreadsBusy"`
}

type RequestInfo struct {
	MaxTime        string `json:"maxTime"`
	ProcessingTime string `json:"processingTime"`
	RequestCount   string `json:"requestCount"`
	ErrorCount     string `json:"errorCount"`
	BytesReceived  string `json:"bytesReceived"`
	BytesSent      string `json:"bytesSent"`
}

type Connector struct {
	Name        string      `json:"name"`
	ThreadInfo  ThreadInfo  `json:"threadInfo"`
	RequestInfo RequestInfo `json:"requestInfo"`
}

// type Manager struct {
// 	ActiveSessions          string `json:"activeSessions"`
// 	SessionCounter          string `json:"sessionCounter"`
// 	MaxActive               string `json:"maxActive"`
// 	RejectedSessions        string `json:"rejectedSessions"`
// 	ExpiredSessions         string `json:"expiredSessions"`
// 	SessionMaxAliveTime     string `json:"sessionMaxAliveTime"`
// 	SessionAverageAliveTime string `json:"sessionAverageAliveTime"`
// 	ProcessingTime          string `json:"processingTime"`
// }

// type Jsp struct {
// 	JspCount       string `json:"jspCount"`
// 	JspReloadCount string `json:"jspReloadCount"`
// }

// type Wrapper struct {
// 	ServletName    string `json:"servletName"`
// 	ProcessingTime string `json:"processingTime"`
// 	MaxTime        string `json:"maxTime"`
// 	RequestCount   string `json:"requestCount"`
// 	ErrorCount     string `json:"errorCount"`
// 	LoadTime       string `json:"loadTime"`
// 	ClassLoadTime  string `json:"classLoadTime"`
// }

// type Context struct {
// 	Name        string    `json:"name"`
// 	StartTime   string    `json:"startTime"`
// 	StartupTime string    `json:"startupTime"`
// 	TldScanTime string    `json:"tldScanTime"`
// 	Manager     Manager   `json:"manager"`
// 	Jsp         Jsp       `json:"jsp"`
// 	Wrapper     []Wrapper `json:"wrapper"`
// }

type TomcatStruct struct {
	TomcatJvm        TomcatJvm   `json:"jvm"`
	TomcatConnectors []Connector `json:"connector"`
	// Context          []Context   `json:"context"`
}
