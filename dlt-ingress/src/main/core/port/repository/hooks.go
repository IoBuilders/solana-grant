package repository

type AfterCreateHook[E any] interface {
	AfterCreate(entity *E) error
}

type AfterDeleteHook[E any] interface {
	AfterDelete(entity *E) error
}

type AfterSaveHook[E any] interface {
	AfterSave(entity *E) error
}

type AfterFindHook[E any] interface {
	AfterFind(entity *E) error
}
