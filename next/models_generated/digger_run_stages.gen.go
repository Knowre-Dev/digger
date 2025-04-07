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

func newDiggerRunStage(db *gorm.DB, opts ...gen.DOOption) diggerRunStage {
	_diggerRunStage := diggerRunStage{}

	_diggerRunStage.diggerRunStageDo.UseDB(db, opts...)
	_diggerRunStage.diggerRunStageDo.UseModel(&model.DiggerRunStage{})

	tableName := _diggerRunStage.diggerRunStageDo.TableName()
	_diggerRunStage.ALL = field.NewAsterisk(tableName)
	_diggerRunStage.ID = field.NewString(tableName, "id")
	_diggerRunStage.CreatedAt = field.NewTime(tableName, "created_at")
	_diggerRunStage.UpdatedAt = field.NewTime(tableName, "updated_at")
	_diggerRunStage.DeletedAt = field.NewField(tableName, "deleted_at")
	_diggerRunStage.BatchID = field.NewString(tableName, "batch_id")

	_diggerRunStage.fillFieldMap()

	return _diggerRunStage
}

type diggerRunStage struct {
	diggerRunStageDo

	ALL       field.Asterisk
	ID        field.String
	CreatedAt field.Time
	UpdatedAt field.Time
	DeletedAt field.Field
	BatchID   field.String

	fieldMap map[string]field.Expr
}

func (d diggerRunStage) Table(newTableName string) *diggerRunStage {
	d.diggerRunStageDo.UseTable(newTableName)
	return d.updateTableName(newTableName)
}

func (d diggerRunStage) As(alias string) *diggerRunStage {
	d.diggerRunStageDo.DO = *(d.diggerRunStageDo.As(alias).(*gen.DO))
	return d.updateTableName(alias)
}

func (d *diggerRunStage) updateTableName(table string) *diggerRunStage {
	d.ALL = field.NewAsterisk(table)
	d.ID = field.NewString(table, "id")
	d.CreatedAt = field.NewTime(table, "created_at")
	d.UpdatedAt = field.NewTime(table, "updated_at")
	d.DeletedAt = field.NewField(table, "deleted_at")
	d.BatchID = field.NewString(table, "batch_id")

	d.fillFieldMap()

	return d
}

func (d *diggerRunStage) GetFieldByName(fieldName string) (field.OrderExpr, bool) {
	_f, ok := d.fieldMap[fieldName]
	if !ok || _f == nil {
		return nil, false
	}
	_oe, ok := _f.(field.OrderExpr)
	return _oe, ok
}

func (d *diggerRunStage) fillFieldMap() {
	d.fieldMap = make(map[string]field.Expr, 5)
	d.fieldMap["id"] = d.ID
	d.fieldMap["created_at"] = d.CreatedAt
	d.fieldMap["updated_at"] = d.UpdatedAt
	d.fieldMap["deleted_at"] = d.DeletedAt
	d.fieldMap["batch_id"] = d.BatchID
}

func (d diggerRunStage) clone(db *gorm.DB) diggerRunStage {
	d.diggerRunStageDo.ReplaceConnPool(db.Statement.ConnPool)
	return d
}

func (d diggerRunStage) replaceDB(db *gorm.DB) diggerRunStage {
	d.diggerRunStageDo.ReplaceDB(db)
	return d
}

type diggerRunStageDo struct{ gen.DO }

type IDiggerRunStageDo interface {
	gen.SubQuery
	Debug() IDiggerRunStageDo
	WithContext(ctx context.Context) IDiggerRunStageDo
	WithResult(fc func(tx gen.Dao)) gen.ResultInfo
	ReplaceDB(db *gorm.DB)
	ReadDB() IDiggerRunStageDo
	WriteDB() IDiggerRunStageDo
	As(alias string) gen.Dao
	Session(config *gorm.Session) IDiggerRunStageDo
	Columns(cols ...field.Expr) gen.Columns
	Clauses(conds ...clause.Expression) IDiggerRunStageDo
	Not(conds ...gen.Condition) IDiggerRunStageDo
	Or(conds ...gen.Condition) IDiggerRunStageDo
	Select(conds ...field.Expr) IDiggerRunStageDo
	Where(conds ...gen.Condition) IDiggerRunStageDo
	Order(conds ...field.Expr) IDiggerRunStageDo
	Distinct(cols ...field.Expr) IDiggerRunStageDo
	Omit(cols ...field.Expr) IDiggerRunStageDo
	Join(table schema.Tabler, on ...field.Expr) IDiggerRunStageDo
	LeftJoin(table schema.Tabler, on ...field.Expr) IDiggerRunStageDo
	RightJoin(table schema.Tabler, on ...field.Expr) IDiggerRunStageDo
	Group(cols ...field.Expr) IDiggerRunStageDo
	Having(conds ...gen.Condition) IDiggerRunStageDo
	Limit(limit int) IDiggerRunStageDo
	Offset(offset int) IDiggerRunStageDo
	Count() (count int64, err error)
	Scopes(funcs ...func(gen.Dao) gen.Dao) IDiggerRunStageDo
	Unscoped() IDiggerRunStageDo
	Create(values ...*model.DiggerRunStage) error
	CreateInBatches(values []*model.DiggerRunStage, batchSize int) error
	Save(values ...*model.DiggerRunStage) error
	First() (*model.DiggerRunStage, error)
	Take() (*model.DiggerRunStage, error)
	Last() (*model.DiggerRunStage, error)
	Find() ([]*model.DiggerRunStage, error)
	FindInBatch(batchSize int, fc func(tx gen.Dao, batch int) error) (results []*model.DiggerRunStage, err error)
	FindInBatches(result *[]*model.DiggerRunStage, batchSize int, fc func(tx gen.Dao, batch int) error) error
	Pluck(column field.Expr, dest interface{}) error
	Delete(...*model.DiggerRunStage) (info gen.ResultInfo, err error)
	Update(column field.Expr, value interface{}) (info gen.ResultInfo, err error)
	UpdateSimple(columns ...field.AssignExpr) (info gen.ResultInfo, err error)
	Updates(value interface{}) (info gen.ResultInfo, err error)
	UpdateColumn(column field.Expr, value interface{}) (info gen.ResultInfo, err error)
	UpdateColumnSimple(columns ...field.AssignExpr) (info gen.ResultInfo, err error)
	UpdateColumns(value interface{}) (info gen.ResultInfo, err error)
	UpdateFrom(q gen.SubQuery) gen.Dao
	Attrs(attrs ...field.AssignExpr) IDiggerRunStageDo
	Assign(attrs ...field.AssignExpr) IDiggerRunStageDo
	Joins(fields ...field.RelationField) IDiggerRunStageDo
	Preload(fields ...field.RelationField) IDiggerRunStageDo
	FirstOrInit() (*model.DiggerRunStage, error)
	FirstOrCreate() (*model.DiggerRunStage, error)
	FindByPage(offset int, limit int) (result []*model.DiggerRunStage, count int64, err error)
	ScanByPage(result interface{}, offset int, limit int) (count int64, err error)
	Scan(result interface{}) (err error)
	Returning(value interface{}, columns ...string) IDiggerRunStageDo
	UnderlyingDB() *gorm.DB
	schema.Tabler
}

func (d diggerRunStageDo) Debug() IDiggerRunStageDo {
	return d.withDO(d.DO.Debug())
}

func (d diggerRunStageDo) WithContext(ctx context.Context) IDiggerRunStageDo {
	return d.withDO(d.DO.WithContext(ctx))
}

func (d diggerRunStageDo) ReadDB() IDiggerRunStageDo {
	return d.Clauses(dbresolver.Read)
}

func (d diggerRunStageDo) WriteDB() IDiggerRunStageDo {
	return d.Clauses(dbresolver.Write)
}

func (d diggerRunStageDo) Session(config *gorm.Session) IDiggerRunStageDo {
	return d.withDO(d.DO.Session(config))
}

func (d diggerRunStageDo) Clauses(conds ...clause.Expression) IDiggerRunStageDo {
	return d.withDO(d.DO.Clauses(conds...))
}

func (d diggerRunStageDo) Returning(value interface{}, columns ...string) IDiggerRunStageDo {
	return d.withDO(d.DO.Returning(value, columns...))
}

func (d diggerRunStageDo) Not(conds ...gen.Condition) IDiggerRunStageDo {
	return d.withDO(d.DO.Not(conds...))
}

func (d diggerRunStageDo) Or(conds ...gen.Condition) IDiggerRunStageDo {
	return d.withDO(d.DO.Or(conds...))
}

func (d diggerRunStageDo) Select(conds ...field.Expr) IDiggerRunStageDo {
	return d.withDO(d.DO.Select(conds...))
}

func (d diggerRunStageDo) Where(conds ...gen.Condition) IDiggerRunStageDo {
	return d.withDO(d.DO.Where(conds...))
}

func (d diggerRunStageDo) Order(conds ...field.Expr) IDiggerRunStageDo {
	return d.withDO(d.DO.Order(conds...))
}

func (d diggerRunStageDo) Distinct(cols ...field.Expr) IDiggerRunStageDo {
	return d.withDO(d.DO.Distinct(cols...))
}

func (d diggerRunStageDo) Omit(cols ...field.Expr) IDiggerRunStageDo {
	return d.withDO(d.DO.Omit(cols...))
}

func (d diggerRunStageDo) Join(table schema.Tabler, on ...field.Expr) IDiggerRunStageDo {
	return d.withDO(d.DO.Join(table, on...))
}

func (d diggerRunStageDo) LeftJoin(table schema.Tabler, on ...field.Expr) IDiggerRunStageDo {
	return d.withDO(d.DO.LeftJoin(table, on...))
}

func (d diggerRunStageDo) RightJoin(table schema.Tabler, on ...field.Expr) IDiggerRunStageDo {
	return d.withDO(d.DO.RightJoin(table, on...))
}

func (d diggerRunStageDo) Group(cols ...field.Expr) IDiggerRunStageDo {
	return d.withDO(d.DO.Group(cols...))
}

func (d diggerRunStageDo) Having(conds ...gen.Condition) IDiggerRunStageDo {
	return d.withDO(d.DO.Having(conds...))
}

func (d diggerRunStageDo) Limit(limit int) IDiggerRunStageDo {
	return d.withDO(d.DO.Limit(limit))
}

func (d diggerRunStageDo) Offset(offset int) IDiggerRunStageDo {
	return d.withDO(d.DO.Offset(offset))
}

func (d diggerRunStageDo) Scopes(funcs ...func(gen.Dao) gen.Dao) IDiggerRunStageDo {
	return d.withDO(d.DO.Scopes(funcs...))
}

func (d diggerRunStageDo) Unscoped() IDiggerRunStageDo {
	return d.withDO(d.DO.Unscoped())
}

func (d diggerRunStageDo) Create(values ...*model.DiggerRunStage) error {
	if len(values) == 0 {
		return nil
	}
	return d.DO.Create(values)
}

func (d diggerRunStageDo) CreateInBatches(values []*model.DiggerRunStage, batchSize int) error {
	return d.DO.CreateInBatches(values, batchSize)
}

// Save : !!! underlying implementation is different with GORM
// The method is equivalent to executing the statement: db.Clauses(clause.OnConflict{UpdateAll: true}).Create(values)
func (d diggerRunStageDo) Save(values ...*model.DiggerRunStage) error {
	if len(values) == 0 {
		return nil
	}
	return d.DO.Save(values)
}

func (d diggerRunStageDo) First() (*model.DiggerRunStage, error) {
	if result, err := d.DO.First(); err != nil {
		return nil, err
	} else {
		return result.(*model.DiggerRunStage), nil
	}
}

func (d diggerRunStageDo) Take() (*model.DiggerRunStage, error) {
	if result, err := d.DO.Take(); err != nil {
		return nil, err
	} else {
		return result.(*model.DiggerRunStage), nil
	}
}

func (d diggerRunStageDo) Last() (*model.DiggerRunStage, error) {
	if result, err := d.DO.Last(); err != nil {
		return nil, err
	} else {
		return result.(*model.DiggerRunStage), nil
	}
}

func (d diggerRunStageDo) Find() ([]*model.DiggerRunStage, error) {
	result, err := d.DO.Find()
	return result.([]*model.DiggerRunStage), err
}

func (d diggerRunStageDo) FindInBatch(batchSize int, fc func(tx gen.Dao, batch int) error) (results []*model.DiggerRunStage, err error) {
	buf := make([]*model.DiggerRunStage, 0, batchSize)
	err = d.DO.FindInBatches(&buf, batchSize, func(tx gen.Dao, batch int) error {
		defer func() { results = append(results, buf...) }()
		return fc(tx, batch)
	})
	return results, err
}

func (d diggerRunStageDo) FindInBatches(result *[]*model.DiggerRunStage, batchSize int, fc func(tx gen.Dao, batch int) error) error {
	return d.DO.FindInBatches(result, batchSize, fc)
}

func (d diggerRunStageDo) Attrs(attrs ...field.AssignExpr) IDiggerRunStageDo {
	return d.withDO(d.DO.Attrs(attrs...))
}

func (d diggerRunStageDo) Assign(attrs ...field.AssignExpr) IDiggerRunStageDo {
	return d.withDO(d.DO.Assign(attrs...))
}

func (d diggerRunStageDo) Joins(fields ...field.RelationField) IDiggerRunStageDo {
	for _, _f := range fields {
		d = *d.withDO(d.DO.Joins(_f))
	}
	return &d
}

func (d diggerRunStageDo) Preload(fields ...field.RelationField) IDiggerRunStageDo {
	for _, _f := range fields {
		d = *d.withDO(d.DO.Preload(_f))
	}
	return &d
}

func (d diggerRunStageDo) FirstOrInit() (*model.DiggerRunStage, error) {
	if result, err := d.DO.FirstOrInit(); err != nil {
		return nil, err
	} else {
		return result.(*model.DiggerRunStage), nil
	}
}

func (d diggerRunStageDo) FirstOrCreate() (*model.DiggerRunStage, error) {
	if result, err := d.DO.FirstOrCreate(); err != nil {
		return nil, err
	} else {
		return result.(*model.DiggerRunStage), nil
	}
}

func (d diggerRunStageDo) FindByPage(offset int, limit int) (result []*model.DiggerRunStage, count int64, err error) {
	result, err = d.Offset(offset).Limit(limit).Find()
	if err != nil {
		return
	}

	if size := len(result); 0 < limit && 0 < size && size < limit {
		count = int64(size + offset)
		return
	}

	count, err = d.Offset(-1).Limit(-1).Count()
	return
}

func (d diggerRunStageDo) ScanByPage(result interface{}, offset int, limit int) (count int64, err error) {
	count, err = d.Count()
	if err != nil {
		return
	}

	err = d.Offset(offset).Limit(limit).Scan(result)
	return
}

func (d diggerRunStageDo) Scan(result interface{}) (err error) {
	return d.DO.Scan(result)
}

func (d diggerRunStageDo) Delete(models ...*model.DiggerRunStage) (result gen.ResultInfo, err error) {
	return d.DO.Delete(models)
}

func (d *diggerRunStageDo) withDO(do gen.Dao) *diggerRunStageDo {
	d.DO = *do.(*gen.DO)
	return d
}
