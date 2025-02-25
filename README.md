# Jamger
在main启动时，和etcd检测key过期时进入protect模式
所有模块（除config、log、global）需要将全局方法改为类方法，以便通过interface传递来解耦