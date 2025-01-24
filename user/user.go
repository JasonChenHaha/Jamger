package juser

type User struct {
	Uid    uint32
	aesKey []byte

	dirty bool
}

// ------------------------- package -------------------------

func newUser(uid uint32, data map[string]string) *User {
	user := &User{Uid: uid}
	if data != nil {
		user.aesKey = []byte(data["aesKey"])
	}
	// jschedule.DoEvery(3*time.Second, user.tick)
	return user
}

// ------------------------- outside -------------------------

func (user *User) SetAesKey(key []byte) {
	user.aesKey = key
	user.dirty = true
}

func (user *User) GetAesKey() []byte {
	return user.aesKey
}

// ------------------------- inside -------------------------

// func (user *User) tick() {
// 	if user.dirty {
// 		if _, err := jdb.Redis.HSet(jglobal.Itoa(user.Uid), "aesKey", user.aesKey); err != nil {
// 			jlog.Error(err)
// 		} else {
// 			user.dirty = false
// 		}
// 	}
// }
