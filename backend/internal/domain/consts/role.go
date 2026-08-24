package consts

// 角色
const (
	RoleSuperAdmin int8 = 1 // 超管
	RoleTeacher    int8 = 2 // 教师
	RoleStudent    int8 = 3 // 学生
)

// RoleText 角色名
func RoleText(r int8) string {
	switch r {
	case RoleSuperAdmin:
		return "超管"
	case RoleTeacher:
		return "教师"
	case RoleStudent:
		return "学生"
	}
	return "未知"
}


// 用户状态
const (
	UserStatusDisabled int8 = 0 // 禁用
	UserStatusNormal   int8 = 1 // 正常
)
