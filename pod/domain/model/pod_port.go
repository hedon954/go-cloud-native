package model

// PodPort Pod 端口相关信息
type PodPort struct {
	ID            int64  `gorm:"primaryKey;notNull;autoIncrement" json:"id"`
	PodID         int64  `json:"pod_id"`
	ContainerPort int32  `json:"container_port"`
	Protocol      string `json:"protocol"` // UPD, TCP
	// TODO: HostPort, etc...
}
