# Jamger
处理gate的顶号问题
    在第二次登录时,codec用的是缓存的旧user.aesKey去解包，导致无法解开
    user.isNew问题，在isNew之前，可能有多次GetUser操作
处理非gate节点缓存数据问题