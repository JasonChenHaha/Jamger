package juser

type Auth struct {
	dm     map[string]any
	dr     map[string]any
	Id     string
	Pwd    []byte
	AesKey []byte
}

// ------------------------- package -------------------------

func newAuth(user *User, mData map[string]any, rData map[string]string) *Auth {
	auth := &Auth{
		dm: user.Base.DirtyMongo,
		dr: user.Base.DirtyRedis,
	}
	if sub, ok := mData["auth"]; ok {
		mData = sub.(map[string]any)
		auth.Id = mData["id"].(string)
		auth.Pwd = mData["pwd"].([]byte)
	}
	if sub, ok := rData["aesKey"]; ok {
		auth.AesKey = []byte(sub)
	}
	return auth
}

// ------------------------- outside -------------------------

func (auth *Auth) SetId(id string) {
	auth.Id = id
	auth.dm["auth.id"] = id
}

func (auth *Auth) SetPwd(pwd []byte) {
	auth.Pwd = pwd
	auth.dm["auth.pwd"] = pwd
}

func (auth *Auth) SetAesKey(aesKey []byte) {
	auth.AesKey = aesKey
	auth.dr["aesKey"] = aesKey
}
