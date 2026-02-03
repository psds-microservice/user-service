package constants

// Разрешения (permissions) по действиям (promt.txt).
const (
	PermStreamCreate     = "stream:create"
	PermStreamJoin       = "stream:join"
	PermChatSend         = "chat:send"
	PermFileUpload       = "file:upload"
	PermConsultationJoin = "consultation:join"
	PermOperatorVerify   = "operator:verify"
	PermOperatorStats    = "operator:stats"
)

// PermissionsByRole — маппинг ролей на списки разрешений.
var PermissionsByRole = map[string][]string{
	RoleClient:   {PermStreamCreate, PermStreamJoin, PermChatSend, PermFileUpload},
	RoleOperator: {PermStreamJoin, PermChatSend, PermFileUpload, PermConsultationJoin},
	RoleAdmin:    {PermStreamCreate, PermStreamJoin, PermChatSend, PermFileUpload, PermConsultationJoin, PermOperatorVerify, PermOperatorStats},
}
