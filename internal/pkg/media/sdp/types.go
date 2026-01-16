package sdp

// SessionDescription represents an SDP session description
type SessionDescription struct {
	Version       int                 // SDP version (always 0)
	Origin        Origin              // Originator and session identifier
	SessionName   string              // Session name
	SessionInfo   string              // Session information
	URI           string              // URI of description
	Emails        []string            // Email addresses
	Phones        []string            // Phone numbers
	Connection    *ConnectionData     // Connection information
	Bandwidth     []Bandwidth         // Zero or more bandwidth information
	Timing        *Timing             // Timing information
	RepeatTimes   []RepeatTime        // Zero or more repeat times
	TimeZones     []TimeZone          // Time zone adjustments
	EncryptionKey string              // Encryption key
	Attributes    []Attribute         // Zero or more session attributes
	MediaDesc     []*MediaDescription // Zero or more media descriptions
}

// Origin contains originator information
type Origin struct {
	Username       string // Username of originator
	SessionID      string // Unique session ID
	SessionVersion string // Session version
	NetType        string // Network type ("IN" for internet)
	AddrType       string // Address type ("IP4" or "IP6")
	UnicastAddr    string // Unicast address
}

// ConnectionData contains connection information
type ConnectionData struct {
	NetType        string // Network type ("IN" for internet)
	AddrType       string // Address type ("IP4" or "IP6")
	ConnectionAddr string // Connection address
	TTL            int    // TTL for multicast
	NumAddr        int    // Number of addresses
}

// Bandwidth contains bandwidth information
type Bandwidth struct {
	BandwidthType string // Type of bandwidth (CT or AS)
	Bandwidth     int    // Bandwidth value in kilobits per second
}

// Timing contains timing information
type Timing struct {
	Start int64 // Session start time (NTP timestamp)
	Stop  int64 // Session stop time (NTP timestamp)
}

// RepeatTime contains repeat times
type RepeatTime struct {
	Intervals []int64 // Zero or more repeat intervals
	Offsets   []int64 // Zero or more offsets from start time
}

// TimeZone contains timezone information
type TimeZone struct {
	AdjustmentTime int64 // NTP timestamp
	Offset         int64 // Offset relative to time in Timing section
}

// Attribute contains attribute information
type Attribute struct {
	Key   string // Attribute name
	Value string // Attribute value (optional)
}

// MediaDescription describes a media stream
type MediaDescription struct {
	MediaName     string          // Media type and transport port
	MediaTitle    string          // Optional media title
	Connection    *ConnectionData // Connection information
	Bandwidth     []Bandwidth     // Zero or more bandwidth information
	EncryptionKey string          // Encryption key
	Attributes    []Attribute     // Zero or more media attributes
}
