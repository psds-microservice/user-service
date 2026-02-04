package constants

// Разрешения (permissions) по действиям. Используются в JWT и при проверках прав в user-service и связанных сервисах.
const (
	PermStreamCreate     = "stream:create"
	PermStreamJoin       = "stream:join"
	PermChatSend         = "chat:send"
	PermFileUpload       = "file:upload"
	PermConsultationJoin = "consultation:join"
	PermOperatorVerify   = "operator:verify"
	PermOperatorStats    = "operator:stats"
)

// PermissionsByRole — маппинг ролей на списки разрешений (домен user-service).
var PermissionsByRole = map[string][]string{
	RoleClient:   {PermStreamCreate, PermStreamJoin, PermChatSend, PermFileUpload},
	RoleOperator: {PermStreamJoin, PermChatSend, PermFileUpload, PermConsultationJoin},
	RoleAdmin:    {PermStreamCreate, PermStreamJoin, PermChatSend, PermFileUpload, PermConsultationJoin, PermOperatorVerify, PermOperatorStats},
}
