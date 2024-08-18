package repository

import "database/sql"

type RepositoryFactoryInterface interface {
	CreteUserRepository(db *sql.DB) UserRepositoryInterface
	CreateFriendRepository(db *sql.DB) FriendRepositoryInterface
	CreatePostRepository(db *sql.DB) PostRepositoryInterface
}

type RepositoryFactory struct{}

func (factory *RepositoryFactory) CreateUserRepository(db *sql.DB) UserRepositoryInterface {
	return &UserRepository{db: db}
}

func (factory *RepositoryFactory) CreatePostRepository(db *sql.DB) PostRepositoryInterface {
	return &PostRepository{db: db}
}
