package manager

import "gorm.io/gorm"

/*
	有些查询，在不同数据库中是存在差异的，但Gorm中并没有进行实现或兼容性实现
	因此，对此类查询在此进行一些封装
	例如：按随机顺序对查询结果排序，在Mysql中为RAND，在SQLite中为RANDOM
	使用方法：db.Scopes(下列封装方法).其它正常调用
	示例：proxy.GetDB().Scopes(proxy.SQLRandomOrder).Take(&user)
*/

// SQLRandomOrder SQL查询结果按随机顺序排序
func (p *PluginProxy) SQLRandomOrder(db *gorm.DB) *gorm.DB {
	switch p.u.dbConfig.Type {
	case MySQL:
		return db.Order("RAND()")
	default:
		return db.Order("RANDOM()")
	}
}
