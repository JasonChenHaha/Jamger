package juser

type User struct {
	Uid    uint32
	AesKey []byte
}

// ------------------------- outside -------------------------

func NewUser(uid uint32) *User {
	return &User{Uid: uid}
}
