package service

import "errors"

var (
	ErrUserNotFound            = errors.New("user not found")
	ErrUserEmailTaken          = errors.New("email already exists")
	ErrUserUsernameTaken       = errors.New("username already exists")
	ErrUserRoleNotFound        = errors.New("role not found")
	ErrCannotDeleteSelf        = errors.New("you cannot delete your own account")
	ErrCannotDemoteSelf        = errors.New("you cannot remove your own super admin role")
	ErrLastSuperAdminProtected = errors.New("cannot remove or demote the last super admin")
)
