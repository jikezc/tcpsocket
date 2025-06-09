package global

import "time"

// MsgExpireTime 消息过期时间，单位秒
const MsgExpireTime int64 = 60

// HeartbeatInterval 心跳间隔时间，单位秒
const HeartbeatInterval time.Duration = 30

// HeartbeatTimeout 心跳超时时间，单位秒
const HeartbeatTimeout int64 = 60

// HeartbeatCheckTime 心跳检查时间，单位秒
const HeartbeatCheckTime time.Duration = 15
