package orm

import "errors"

const (
	keywordAutoIncrement = "autoIncrement"
	keywordPrimaryKey    = "primaryKey"
	keywordUniqueKey     = "uniqueKey"
	keywordIgnoreField   = "-"
)

const (
	logicalAnd = "AND"
	logicalOr  = "OR"

	operateEquals = "="
	// operateNotEquals          = "!="
	// operateLessThan           = "<"
	// operateLessThanOrEqual    = "<="
	// operateGreaterThan        = ">"
	// operateGreaterThanOrEqual = ">="
	operateIn         = "IN"
	operateNotIn      = "NOT IN"
	operateIs         = "IS"
	operateIsNot      = "IS NOT"
	operateOrderByAsc = "ASC"
	// operateOrderByDesc        = "DESC"

	bracketOpen  = "("
	bracketClose = ")"
)

var (
	// ErrRecordNotFound 错误：记录不存在
	ErrRecordNotFound = errors.New("record not found")

	// ErrMissingModel 错误：未继承 ORM Model
	ErrMissingModel = errors.New("missing model")

	// ErrUnknownStruct 错误：无法解析的结构体
	ErrUnknownStruct = errors.New("unknown struct")

	// ErrSQLEmpty 错误：SQL 为空
	ErrSQLEmpty = errors.New("sql is empty")

	// ErrNeedPointer 错误：必须是指针
	ErrNeedPointer = errors.New("need a pointer")

	// ErrElementNeedStruct 错误：必须是个结构体
	ErrElementNeedStruct = errors.New("element need a struct")

	// ErrNeedPtrToSlice 错误：必须是指针数组
	ErrNeedPtrToSlice = errors.New("need a pointer to a slice")

	// ErrNoPrimaryKey 错误：没有设置主键
	ErrNoPrimaryKey = errors.New("not primary key")

	// ErrNoPrimaryAndUnique 错误：没有主键和唯一键
	ErrNoPrimaryAndUnique = errors.New("not primary key and unique key")

	// ErrUpdateParamsEmpty 错误：更新参数为空
	ErrUpdateParamsEmpty = errors.New("update params empty")

	// ErrDuplicateValues 错误：重复赋值
	ErrDuplicateValues = errors.New("duplicate values")

	// ErrFieldsNotMatch 错误：字段不匹配
	ErrFieldsNotMatch = errors.New("fields not match")

	// ErrNotSetUpdateField 错误：未设置更新字段
	ErrNotSetUpdateField = errors.New("not set update field")

	// ErrTransExist 错误：事物未关闭
	ErrTransExist = errors.New("transaction has exist")

	// ErrTransNotExist 错误：事物不存在
	ErrTransNotExist = errors.New("transaction not exist")

	// ErrFromEmpty 错误: 没有 FROM 条件
	ErrFromEmpty = errors.New("from empty")

	// ErrWhereEmpty 错误: 没有 WHERE 条件
	ErrWhereEmpty = errors.New("where empty")

	// ErrSetField 错误: 设置更新字段错误
	ErrSetField = errors.New("unsupport set field")

	// ErrNotSetInsertField 错误: 未设置插入字段
	ErrNotSetInsertField = errors.New("not set insert field")
)
