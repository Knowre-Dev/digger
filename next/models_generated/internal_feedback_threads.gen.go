// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.

package models_generated

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"

	"gorm.io/gen"
	"gorm.io/gen/field"

	"gorm.io/plugin/dbresolver"

	"github.com/diggerhq/digger/next/model"
)

func newInternalFeedbackThread(db *gorm.DB, opts ...gen.DOOption) internalFeedbackThread {
	_internalFeedbackThread := internalFeedbackThread{}

	_internalFeedbackThread.internalFeedbackThreadDo.UseDB(db, opts...)
	_internalFeedbackThread.internalFeedbackThreadDo.UseModel(&model.InternalFeedbackThread{})

	tableName := _internalFeedbackThread.internalFeedbackThreadDo.TableName()
	_internalFeedbackThread.ALL = field.NewAsterisk(tableName)
	_internalFeedbackThread.ID = field.NewString(tableName, "id")
	_internalFeedbackThread.Title = field.NewString(tableName, "title")
	_internalFeedbackThread.Content = field.NewString(tableName, "content")
	_internalFeedbackThread.UserID = field.NewString(tableName, "user_id")
	_internalFeedbackThread.CreatedAt = field.NewTime(tableName, "created_at")
	_internalFeedbackThread.UpdatedAt = field.NewTime(tableName, "updated_at")
	_internalFeedbackThread.Priority = field.NewString(tableName, "priority")
	_internalFeedbackThread.Type = field.NewString(tableName, "type")
	_internalFeedbackThread.Status = field.NewString(tableName, "status")
	_internalFeedbackThread.AddedToRoadmap = field.NewBool(tableName, "added_to_roadmap")
	_internalFeedbackThread.OpenForPublicDiscussion = field.NewBool(tableName, "open_for_public_discussion")
	_internalFeedbackThread.IsPubliclyVisible = field.NewBool(tableName, "is_publicly_visible")

	_internalFeedbackThread.fillFieldMap()

	return _internalFeedbackThread
}

type internalFeedbackThread struct {
	internalFeedbackThreadDo

	ALL                     field.Asterisk
	ID                      field.String
	Title                   field.String
	Content                 field.String
	UserID                  field.String
	CreatedAt               field.Time
	UpdatedAt               field.Time
	Priority                field.String
	Type                    field.String
	Status                  field.String
	AddedToRoadmap          field.Bool
	OpenForPublicDiscussion field.Bool
	IsPubliclyVisible       field.Bool

	fieldMap map[string]field.Expr
}

func (i internalFeedbackThread) Table(newTableName string) *internalFeedbackThread {
	i.internalFeedbackThreadDo.UseTable(newTableName)
	return i.updateTableName(newTableName)
}

func (i internalFeedbackThread) As(alias string) *internalFeedbackThread {
	i.internalFeedbackThreadDo.DO = *(i.internalFeedbackThreadDo.As(alias).(*gen.DO))
	return i.updateTableName(alias)
}

func (i *internalFeedbackThread) updateTableName(table string) *internalFeedbackThread {
	i.ALL = field.NewAsterisk(table)
	i.ID = field.NewString(table, "id")
	i.Title = field.NewString(table, "title")
	i.Content = field.NewString(table, "content")
	i.UserID = field.NewString(table, "user_id")
	i.CreatedAt = field.NewTime(table, "created_at")
	i.UpdatedAt = field.NewTime(table, "updated_at")
	i.Priority = field.NewString(table, "priority")
	i.Type = field.NewString(table, "type")
	i.Status = field.NewString(table, "status")
	i.AddedToRoadmap = field.NewBool(table, "added_to_roadmap")
	i.OpenForPublicDiscussion = field.NewBool(table, "open_for_public_discussion")
	i.IsPubliclyVisible = field.NewBool(table, "is_publicly_visible")

	i.fillFieldMap()

	return i
}

func (i *internalFeedbackThread) GetFieldByName(fieldName string) (field.OrderExpr, bool) {
	_f, ok := i.fieldMap[fieldName]
	if !ok || _f == nil {
		return nil, false
	}
	_oe, ok := _f.(field.OrderExpr)
	return _oe, ok
}

func (i *internalFeedbackThread) fillFieldMap() {
	i.fieldMap = make(map[string]field.Expr, 12)
	i.fieldMap["id"] = i.ID
	i.fieldMap["title"] = i.Title
	i.fieldMap["content"] = i.Content
	i.fieldMap["user_id"] = i.UserID
	i.fieldMap["created_at"] = i.CreatedAt
	i.fieldMap["updated_at"] = i.UpdatedAt
	i.fieldMap["priority"] = i.Priority
	i.fieldMap["type"] = i.Type
	i.fieldMap["status"] = i.Status
	i.fieldMap["added_to_roadmap"] = i.AddedToRoadmap
	i.fieldMap["open_for_public_discussion"] = i.OpenForPublicDiscussion
	i.fieldMap["is_publicly_visible"] = i.IsPubliclyVisible
}

func (i internalFeedbackThread) clone(db *gorm.DB) internalFeedbackThread {
	i.internalFeedbackThreadDo.ReplaceConnPool(db.Statement.ConnPool)
	return i
}

func (i internalFeedbackThread) replaceDB(db *gorm.DB) internalFeedbackThread {
	i.internalFeedbackThreadDo.ReplaceDB(db)
	return i
}

type internalFeedbackThreadDo struct{ gen.DO }

type IInternalFeedbackThreadDo interface {
	gen.SubQuery
	Debug() IInternalFeedbackThreadDo
	WithContext(ctx context.Context) IInternalFeedbackThreadDo
	WithResult(fc func(tx gen.Dao)) gen.ResultInfo
	ReplaceDB(db *gorm.DB)
	ReadDB() IInternalFeedbackThreadDo
	WriteDB() IInternalFeedbackThreadDo
	As(alias string) gen.Dao
	Session(config *gorm.Session) IInternalFeedbackThreadDo
	Columns(cols ...field.Expr) gen.Columns
	Clauses(conds ...clause.Expression) IInternalFeedbackThreadDo
	Not(conds ...gen.Condition) IInternalFeedbackThreadDo
	Or(conds ...gen.Condition) IInternalFeedbackThreadDo
	Select(conds ...field.Expr) IInternalFeedbackThreadDo
	Where(conds ...gen.Condition) IInternalFeedbackThreadDo
	Order(conds ...field.Expr) IInternalFeedbackThreadDo
	Distinct(cols ...field.Expr) IInternalFeedbackThreadDo
	Omit(cols ...field.Expr) IInternalFeedbackThreadDo
	Join(table schema.Tabler, on ...field.Expr) IInternalFeedbackThreadDo
	LeftJoin(table schema.Tabler, on ...field.Expr) IInternalFeedbackThreadDo
	RightJoin(table schema.Tabler, on ...field.Expr) IInternalFeedbackThreadDo
	Group(cols ...field.Expr) IInternalFeedbackThreadDo
	Having(conds ...gen.Condition) IInternalFeedbackThreadDo
	Limit(limit int) IInternalFeedbackThreadDo
	Offset(offset int) IInternalFeedbackThreadDo
	Count() (count int64, err error)
	Scopes(funcs ...func(gen.Dao) gen.Dao) IInternalFeedbackThreadDo
	Unscoped() IInternalFeedbackThreadDo
	Create(values ...*model.InternalFeedbackThread) error
	CreateInBatches(values []*model.InternalFeedbackThread, batchSize int) error
	Save(values ...*model.InternalFeedbackThread) error
	First() (*model.InternalFeedbackThread, error)
	Take() (*model.InternalFeedbackThread, error)
	Last() (*model.InternalFeedbackThread, error)
	Find() ([]*model.InternalFeedbackThread, error)
	FindInBatch(batchSize int, fc func(tx gen.Dao, batch int) error) (results []*model.InternalFeedbackThread, err error)
	FindInBatches(result *[]*model.InternalFeedbackThread, batchSize int, fc func(tx gen.Dao, batch int) error) error
	Pluck(column field.Expr, dest interface{}) error
	Delete(...*model.InternalFeedbackThread) (info gen.ResultInfo, err error)
	Update(column field.Expr, value interface{}) (info gen.ResultInfo, err error)
	UpdateSimple(columns ...field.AssignExpr) (info gen.ResultInfo, err error)
	Updates(value interface{}) (info gen.ResultInfo, err error)
	UpdateColumn(column field.Expr, value interface{}) (info gen.ResultInfo, err error)
	UpdateColumnSimple(columns ...field.AssignExpr) (info gen.ResultInfo, err error)
	UpdateColumns(value interface{}) (info gen.ResultInfo, err error)
	UpdateFrom(q gen.SubQuery) gen.Dao
	Attrs(attrs ...field.AssignExpr) IInternalFeedbackThreadDo
	Assign(attrs ...field.AssignExpr) IInternalFeedbackThreadDo
	Joins(fields ...field.RelationField) IInternalFeedbackThreadDo
	Preload(fields ...field.RelationField) IInternalFeedbackThreadDo
	FirstOrInit() (*model.InternalFeedbackThread, error)
	FirstOrCreate() (*model.InternalFeedbackThread, error)
	FindByPage(offset int, limit int) (result []*model.InternalFeedbackThread, count int64, err error)
	ScanByPage(result interface{}, offset int, limit int) (count int64, err error)
	Scan(result interface{}) (err error)
	Returning(value interface{}, columns ...string) IInternalFeedbackThreadDo
	UnderlyingDB() *gorm.DB
	schema.Tabler
}

func (i internalFeedbackThreadDo) Debug() IInternalFeedbackThreadDo {
	return i.withDO(i.DO.Debug())
}

func (i internalFeedbackThreadDo) WithContext(ctx context.Context) IInternalFeedbackThreadDo {
	return i.withDO(i.DO.WithContext(ctx))
}

func (i internalFeedbackThreadDo) ReadDB() IInternalFeedbackThreadDo {
	return i.Clauses(dbresolver.Read)
}

func (i internalFeedbackThreadDo) WriteDB() IInternalFeedbackThreadDo {
	return i.Clauses(dbresolver.Write)
}

func (i internalFeedbackThreadDo) Session(config *gorm.Session) IInternalFeedbackThreadDo {
	return i.withDO(i.DO.Session(config))
}

func (i internalFeedbackThreadDo) Clauses(conds ...clause.Expression) IInternalFeedbackThreadDo {
	return i.withDO(i.DO.Clauses(conds...))
}

func (i internalFeedbackThreadDo) Returning(value interface{}, columns ...string) IInternalFeedbackThreadDo {
	return i.withDO(i.DO.Returning(value, columns...))
}

func (i internalFeedbackThreadDo) Not(conds ...gen.Condition) IInternalFeedbackThreadDo {
	return i.withDO(i.DO.Not(conds...))
}

func (i internalFeedbackThreadDo) Or(conds ...gen.Condition) IInternalFeedbackThreadDo {
	return i.withDO(i.DO.Or(conds...))
}

func (i internalFeedbackThreadDo) Select(conds ...field.Expr) IInternalFeedbackThreadDo {
	return i.withDO(i.DO.Select(conds...))
}

func (i internalFeedbackThreadDo) Where(conds ...gen.Condition) IInternalFeedbackThreadDo {
	return i.withDO(i.DO.Where(conds...))
}

func (i internalFeedbackThreadDo) Order(conds ...field.Expr) IInternalFeedbackThreadDo {
	return i.withDO(i.DO.Order(conds...))
}

func (i internalFeedbackThreadDo) Distinct(cols ...field.Expr) IInternalFeedbackThreadDo {
	return i.withDO(i.DO.Distinct(cols...))
}

func (i internalFeedbackThreadDo) Omit(cols ...field.Expr) IInternalFeedbackThreadDo {
	return i.withDO(i.DO.Omit(cols...))
}

func (i internalFeedbackThreadDo) Join(table schema.Tabler, on ...field.Expr) IInternalFeedbackThreadDo {
	return i.withDO(i.DO.Join(table, on...))
}

func (i internalFeedbackThreadDo) LeftJoin(table schema.Tabler, on ...field.Expr) IInternalFeedbackThreadDo {
	return i.withDO(i.DO.LeftJoin(table, on...))
}

func (i internalFeedbackThreadDo) RightJoin(table schema.Tabler, on ...field.Expr) IInternalFeedbackThreadDo {
	return i.withDO(i.DO.RightJoin(table, on...))
}

func (i internalFeedbackThreadDo) Group(cols ...field.Expr) IInternalFeedbackThreadDo {
	return i.withDO(i.DO.Group(cols...))
}

func (i internalFeedbackThreadDo) Having(conds ...gen.Condition) IInternalFeedbackThreadDo {
	return i.withDO(i.DO.Having(conds...))
}

func (i internalFeedbackThreadDo) Limit(limit int) IInternalFeedbackThreadDo {
	return i.withDO(i.DO.Limit(limit))
}

func (i internalFeedbackThreadDo) Offset(offset int) IInternalFeedbackThreadDo {
	return i.withDO(i.DO.Offset(offset))
}

func (i internalFeedbackThreadDo) Scopes(funcs ...func(gen.Dao) gen.Dao) IInternalFeedbackThreadDo {
	return i.withDO(i.DO.Scopes(funcs...))
}

func (i internalFeedbackThreadDo) Unscoped() IInternalFeedbackThreadDo {
	return i.withDO(i.DO.Unscoped())
}

func (i internalFeedbackThreadDo) Create(values ...*model.InternalFeedbackThread) error {
	if len(values) == 0 {
		return nil
	}
	return i.DO.Create(values)
}

func (i internalFeedbackThreadDo) CreateInBatches(values []*model.InternalFeedbackThread, batchSize int) error {
	return i.DO.CreateInBatches(values, batchSize)
}

// Save : !!! underlying implementation is different with GORM
// The method is equivalent to executing the statement: db.Clauses(clause.OnConflict{UpdateAll: true}).Create(values)
func (i internalFeedbackThreadDo) Save(values ...*model.InternalFeedbackThread) error {
	if len(values) == 0 {
		return nil
	}
	return i.DO.Save(values)
}

func (i internalFeedbackThreadDo) First() (*model.InternalFeedbackThread, error) {
	if result, err := i.DO.First(); err != nil {
		return nil, err
	} else {
		return result.(*model.InternalFeedbackThread), nil
	}
}

func (i internalFeedbackThreadDo) Take() (*model.InternalFeedbackThread, error) {
	if result, err := i.DO.Take(); err != nil {
		return nil, err
	} else {
		return result.(*model.InternalFeedbackThread), nil
	}
}

func (i internalFeedbackThreadDo) Last() (*model.InternalFeedbackThread, error) {
	if result, err := i.DO.Last(); err != nil {
		return nil, err
	} else {
		return result.(*model.InternalFeedbackThread), nil
	}
}

func (i internalFeedbackThreadDo) Find() ([]*model.InternalFeedbackThread, error) {
	result, err := i.DO.Find()
	return result.([]*model.InternalFeedbackThread), err
}

func (i internalFeedbackThreadDo) FindInBatch(batchSize int, fc func(tx gen.Dao, batch int) error) (results []*model.InternalFeedbackThread, err error) {
	buf := make([]*model.InternalFeedbackThread, 0, batchSize)
	err = i.DO.FindInBatches(&buf, batchSize, func(tx gen.Dao, batch int) error {
		defer func() { results = append(results, buf...) }()
		return fc(tx, batch)
	})
	return results, err
}

func (i internalFeedbackThreadDo) FindInBatches(result *[]*model.InternalFeedbackThread, batchSize int, fc func(tx gen.Dao, batch int) error) error {
	return i.DO.FindInBatches(result, batchSize, fc)
}

func (i internalFeedbackThreadDo) Attrs(attrs ...field.AssignExpr) IInternalFeedbackThreadDo {
	return i.withDO(i.DO.Attrs(attrs...))
}

func (i internalFeedbackThreadDo) Assign(attrs ...field.AssignExpr) IInternalFeedbackThreadDo {
	return i.withDO(i.DO.Assign(attrs...))
}

func (i internalFeedbackThreadDo) Joins(fields ...field.RelationField) IInternalFeedbackThreadDo {
	for _, _f := range fields {
		i = *i.withDO(i.DO.Joins(_f))
	}
	return &i
}

func (i internalFeedbackThreadDo) Preload(fields ...field.RelationField) IInternalFeedbackThreadDo {
	for _, _f := range fields {
		i = *i.withDO(i.DO.Preload(_f))
	}
	return &i
}

func (i internalFeedbackThreadDo) FirstOrInit() (*model.InternalFeedbackThread, error) {
	if result, err := i.DO.FirstOrInit(); err != nil {
		return nil, err
	} else {
		return result.(*model.InternalFeedbackThread), nil
	}
}

func (i internalFeedbackThreadDo) FirstOrCreate() (*model.InternalFeedbackThread, error) {
	if result, err := i.DO.FirstOrCreate(); err != nil {
		return nil, err
	} else {
		return result.(*model.InternalFeedbackThread), nil
	}
}

func (i internalFeedbackThreadDo) FindByPage(offset int, limit int) (result []*model.InternalFeedbackThread, count int64, err error) {
	result, err = i.Offset(offset).Limit(limit).Find()
	if err != nil {
		return
	}

	if size := len(result); 0 < limit && 0 < size && size < limit {
		count = int64(size + offset)
		return
	}

	count, err = i.Offset(-1).Limit(-1).Count()
	return
}

func (i internalFeedbackThreadDo) ScanByPage(result interface{}, offset int, limit int) (count int64, err error) {
	count, err = i.Count()
	if err != nil {
		return
	}

	err = i.Offset(offset).Limit(limit).Scan(result)
	return
}

func (i internalFeedbackThreadDo) Scan(result interface{}) (err error) {
	return i.DO.Scan(result)
}

func (i internalFeedbackThreadDo) Delete(models ...*model.InternalFeedbackThread) (result gen.ResultInfo, err error) {
	return i.DO.Delete(models)
}

func (i *internalFeedbackThreadDo) withDO(do gen.Dao) *internalFeedbackThreadDo {
	i.DO = *do.(*gen.DO)
	return i
}
