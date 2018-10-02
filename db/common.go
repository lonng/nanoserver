package db

import (
	"fmt"
	"strings"
)

type Option func(setting *Setting)
type Closer func()

type Setting struct {
	ShowSQL      bool
	MaxOpenConns int
	MaxIdleConns int
}

func MaxIdleConnOption(i int) Option {
	return func(s *Setting) {
		s.MaxIdleConns = i
	}
}

func MaxOpenConnOption(i int) Option {
	return func(s *Setting) {
		s.MaxOpenConns = i
	}
}

func ShowSQLOption(show bool) Option {
	return func(s *Setting) {
		s.ShowSQL = show
	}
}

// Build data source name
func BuildDSN(host string, port int, username, password, dbname, args string) string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?%s", username, password, host, port, dbname, args)
}

//// 根据一个大区信息, 返回大区的SQL条件语句
//// 例如: area: {AreaID: 100, ServerIDs: []int{1, 2, 3, 5}}
//// 返回: (`area_id`=100 AND `server_id` IN(1,2,3,5))
//func AreaCondition(area *protocol.Area) string {
//	if len(area.ServerIDs) > 0 {
//		servers := []string{}
//		for _, id := range area.ServerIDs {
//			servers = append(servers, strconv.Itoa(id))
//		}
//		return fmt.Sprintf("(`area_id`=%d AND `server_id` IN(%s))", area.AreaID, strings.Join(servers, ","))
//	} else {
//		return fmt.Sprintf("`area_id`=%d", area.AreaID)
//	}
//}
//
//func AreaVectorCondition(areas []*protocol.Area) string {
//	conditions := []string{}
//	for _, area := range areas {
//		conditions = append(conditions, AreaCondition(area))
//	}
//	return fmt.Sprintf("(%s)", strings.Join(conditions, " OR "))
//}

// 给定列, 返回起始时间条件SQL语句, [begin, end)
func RangeCondition(column string, begin, end int64) string {
	return fmt.Sprintf("(`%s` BETWEEN %d AND %d)", column, begin, end)
}

func ChannelCondition(c []string) string {
	return fmt.Sprintf("`channel` IN('%s')", strings.Join(c, "','"))
}

func EqIntCondition(col string, v int) string {
	return fmt.Sprintf("`%s`=%d", col, v)
}

func EqInt64Condition(col string, v int64) string {
	return fmt.Sprintf("`%s`=%d", col, v)
}

func LtInt64Condition(col string, v int64) string {
	return fmt.Sprintf("`%s`<%d", col, v)
}

func Combined(cond ...string) string {
	return strings.Join(cond, " AND ")
}

func Insert(bean interface{}) error {
	_, err := database.Insert(bean)
	return err
}
