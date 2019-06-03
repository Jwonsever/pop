package pop

import (
	"reflect"
	"testing"

	"github.com/gobuffalo/pop/nulls"
	"github.com/stretchr/testify/require"
)

func Test_Model_Meta(t *testing.T) {
	r := require.New(t)

	m := Model{Value: &User{}}
	mm := m.Meta()

	r.Equal(mm.Type, reflect.TypeOf(m.Value))
	r.Equal(mm.IndirectType, reflect.Indirect(reflect.ValueOf(m.Value)).Type())
}

func Test_Model_Meta_Slice(t *testing.T) {
	r := require.New(t)

	m := Model{Value: &User{}}
	mm := m.Meta()
	sl := mm.MakeSlice()

	r.Equal(sl.IndirectType.Kind(), reflect.Slice)
	r.Equal(sl.IndirectValue.Len(), 0)
}

func Test_Model_Meta_Map_For_Struct(t *testing.T) {
	r := require.New(t)

	u := User{}
	m := Model{Value: &u}
	mm := m.Meta()
	sl := mm.MakeMap()

	r.Equal(sl.Type.Kind(), reflect.Map)
	r.Equal(sl.Value.Type(), reflect.MapOf(reflect.TypeOf(u.ID), mm.Type))
}

func Test_Model_Meta_Map_For_Slice(t *testing.T) {
	r := require.New(t)

	u := []*User{
		{Email: "User@email.com"},
	}
	m := Model{Value: &u}
	mm := m.Meta()
	sl := mm.MakeMap()

	r.Equal(sl.Type, reflect.MapOf(reflect.TypeOf(0), reflect.TypeOf(&User{})))
	r.Equal(1, len(sl.Value.Interface().(map[int]*User)))

	// Map for non-struct with pointer.
	n := 1
	v := []*int{&n}
	m = Model{Value: &v}
	mm = m.Meta()
	sl = mm.MakeMap()
	r.Equal(sl.Type, reflect.MapOf(reflect.TypeOf(""), reflect.TypeOf(v[0])))
	r.Equal(1, len(sl.Value.Interface().(map[string]*int)))

	// Map for non-struct without pointer.
	v2 := []int{1}
	m = Model{Value: &v2}
	mm = m.Meta()
	sl = mm.MakeMap()
	r.Equal(sl.Type, reflect.MapOf(reflect.TypeOf(""), reflect.PtrTo(reflect.TypeOf(v2[0]))))
	r.Equal(1, len(sl.Value.Interface().(map[string]*int)))
}

func Test_Model_Meta_Associations(t *testing.T) {
	r := require.New(t)

	m := Model{Value: &User{}}
	mm := m.Meta()

	mAssociations := mm.Associations()
	r.Equal(3, len(mAssociations))
}

func Test_Model_Meta_Associations_Loading(t *testing.T) {
	transaction(func(tx *Connection) {
		a := require.New(t)

		for _, name := range []string{"Mark", "Joe", "Jane"} {
			User := User{Name: nulls.NewString(name)}
			err := tx.Create(&User)
			a.NoError(err)

			book := Book{UserID: nulls.NewInt(User.ID)}
			err = tx.Create(&book)
			a.NoError(err)

			if name == "Mark" {
				song := Song{UserID: User.ID}
				err = tx.Create(&song)
				a.NoError(err)

				address := Address{Street: "Pop"}
				err = tx.Create(&address)

				home := UsersAddress{UserID: User.ID, AddressID: address.ID}
				err = tx.Create(&home)
			}
		}

		Users := Users{}
		tx.All(&Users)

		mt := (&Model{Value: &Users}).Meta()
		err := mt.LoadDirect(tx, "has_many")
		a.NoError(err)

		err = mt.LoadDirect(tx, "has_one")
		a.NoError(err)

		// err = LoadBidirect(&Users, tx, "many_to_many")
		// a.NoError(err)

		a.Equal(1, len(Users[0].Books))
		a.Equal(Users[0].ID, Users[0].FavoriteSong.UserID)
		a.Zero(Users[1].FavoriteSong.UserID)

		books := Books{}
		err = tx.All(&books)
		a.NoError(err)

		mt = (&Model{Value: &books}).Meta()
		mt.LoadIndirect(tx, "belongs_to")
		a.Equal(Users[0].ID, books[0].User.ID)
		a.Equal(Users[1].ID, books[1].User.ID)
		a.Equal(Users[2].ID, books[2].User.ID)
	})
}
