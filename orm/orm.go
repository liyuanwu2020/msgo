package orm

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/liyuanwu2020/msgo/mslog"
	"reflect"
	"strings"
	"time"
)

type MsDb struct {
	db     *sql.DB
	logger *mslog.Logger
	Prefix string
}

type MsSession struct {
	db          *MsDb
	TableName   string
	fieldName   []string
	placeHolder []string
	values      []any
}

func (s *MsSession) Table(name string) *MsSession {
	s.TableName = name
	return s
}

func (s *MsSession) Insert(data any) (int64, int64, error) {
	s.fieldNames(data)
	query := fmt.Sprintf("insert into %s (%s) values(%s)", s.TableName, strings.Join(s.fieldName, ","), strings.Join(s.placeHolder, ","))
	stmt, err := s.db.db.Prepare(query)
	if err != nil {
		s.db.logger.Error(err)
		return -1, -1, err
	}
	r, err := stmt.Exec(s.values...)
	if err != nil {
		s.db.logger.Error(err)
		return -1, -1, err
	}
	id, err := r.LastInsertId()
	if err != nil {
		s.db.logger.Error(err)
		return -1, -1, err
	}
	affected, err := r.RowsAffected()
	if err != nil {
		s.db.logger.Error(err)
		return -1, -1, err
	}
	return id, affected, nil
}

func (s *MsSession) fieldNames(data any) {
	t := reflect.TypeOf(data)
	v := reflect.ValueOf(data)
	if t.Kind() != reflect.Pointer {
		panic(errors.New("data type must be pointer"))
	}
	tVar := t.Elem()
	vVar := v.Elem()
	if s.TableName == "" {
		s.TableName = s.db.Prefix + strings.ToLower(Name(tVar.Name()))
	}

	var fieldNames []string
	var placeholder []string
	var values []any
	for i := 0; i < tVar.NumField(); i++ {
		//首字母是小写的
		if !vVar.Field(i).CanInterface() {
			continue
		}
		//解析tag
		field := tVar.Field(i)
		sqlTag := field.Tag.Get("mssql")
		if sqlTag == "" {
			sqlTag = strings.ToLower(Name(field.Name))
		}
		contains := strings.Contains(sqlTag, "auto_increment")
		if sqlTag == "id" || contains {
			//对id做个判断 如果其值小于等于0 数据库可能是自增 跳过此字段
			if isAutoId(vVar.Field(i).Interface()) {
				continue
			}
		}
		if contains {
			sqlTag = sqlTag[:strings.Index(sqlTag, ",")]
		}
		fieldNames = append(fieldNames, sqlTag)
		placeholder = append(placeholder, "?")
		values = append(values, vVar.Field(i).Interface())
	}
	s.fieldName = fieldNames
	s.placeHolder = placeholder
	s.values = values
}

func Name(name string) string {
	all := name[:]
	var sb strings.Builder
	lastIndex := 0
	for index, value := range all {
		if value >= 65 && value <= 90 {
			if index == 0 {
				continue
			}
			sb.WriteString(name[lastIndex:index])
			sb.WriteString("_")
			lastIndex = index
		}
	}
	if lastIndex != len(name)-1 {
		sb.WriteString(name[lastIndex:])
	}
	return sb.String()
}

func isAutoId(id any) bool {
	t := reflect.TypeOf(id)
	v := reflect.ValueOf(id)
	switch t.Kind() {
	case reflect.Int64:
		if v.Interface().(int64) <= 0 {
			return true
		}
	case reflect.Int32:
		if v.Interface().(int32) <= 0 {
			return true
		}
	case reflect.Int:
		if v.Interface().(int) <= 0 {
			return true
		}
	default:
		return false
	}
	return false
}

func (d *MsDb) TablePrefix(prefix string) *MsDb {
	d.Prefix = prefix
	return d
}

func (d *MsDb) New() *MsSession {
	return &MsSession{
		db: d,
	}
}

func Open(driverName string, source string) (*MsDb, error) {
	db, err := sql.Open(driverName, source)
	if err != nil {
		panic(err)
	}
	db.SetMaxOpenConns(1)
	msDb := &MsDb{
		db: db,
	}
	//最大空闲连接数，默认不配置，是2个最大空闲连接
	db.SetMaxIdleConns(5)
	//最大连接数，默认不配置，是不限制最大连接数
	db.SetMaxOpenConns(100)
	// 连接最大存活时间
	db.SetConnMaxLifetime(time.Minute * 3)
	//空闲连接最大存活时间
	db.SetConnMaxIdleTime(time.Minute * 1)
	err = db.Ping()
	if err != nil {
		panic(err)
	}
	return msDb, nil
}
