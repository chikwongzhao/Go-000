# 第二课 Go语言实践 - error

## 问题

我们在数据库操作的时候，比如`dao`层中遇到一个`sql.ErrNoRows`的时候，是否应该`Wrap`这个`error`，抛给上层。为什么，应该怎么做请写出代码？

## 个人解答

先以`sql.ErrNoRows`为例，我们先搞清楚为什么在没有数据的时候底层会返回这样一个错误：

以`gorm`为例，源码中的`ErrRecordNotFound`对应有这样一行注释：
```go
// ErrRecordNotFound record not found error, happens when haven't find any matched data when looking up with a struct
ErrRecordNotFound = errors.New("record not found")
```

即：当使用一个`struct`查找记录时找不到数据时会返回此`ErrRecordNotFound`。

再撸源码发现：

```go
} else if scope.db.RowsAffected == 0 && !isSlice {
	scope.Err(ErrRecordNotFound)
}
```

这里注意条件中的`&& !isSlice`条件，即：只有在不使用切片查询的时候才会返回此错误。

那么为什么在使用切片时查询不会返回此错误，但是使用`struct`的时候会返回此错误呢？

设想这样一个case：

当使用一个`struct`想从DB查询某一条数据的时候，假如欲查找的记录不存在，并且`gorm`不会返回此`ErrRecordNotFound`错误，那么我们用于接收结果记录的`struct`中的字段将都为初始化的零值，那么我们在`dao`中如何知道这些零值是`DB`中未找到记录还是`DB`中的数据就是都是零值呢？

所以，`gorm`要返回一个错误明确告诉调用方找不到记录。

那么，搞清楚了这个，我们就很容易知道我们的`dao`层要不要往上抛此错误了：

不需要。

因为我们在`dao`中已经明确知道了没有找到记录，而不是`DB`发生了什么错误，所以需要在`dao`中吞掉此错误，返回一个数据`nil`和一个错误`nil`给业务调用方即可。

```go
func QueryXXX(param XXX) (data *XXXX, err error) {
	err := sql.Get(databaseName).Slave().Table(tableName).Where(xxx).Find(m).Error
	if gorm.IsRecordNotFoundError(err) {
		return nil, nil
	}
	// do something else
}
```

