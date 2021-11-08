package service

import (
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"pc_book/pd"
)

// AuthService is a the server for authentication
type AuthService struct {
	pd.UnimplementedAuthServiceServer
	userStore  UserStore
	jwtManager *JWTManager
}

// NewAuthService return a new auth server
func NewAuthService(userStore UserStore, jwtManager *JWTManager) *AuthService {
	return &AuthService{
		userStore: userStore,
		jwtManager: jwtManager,
	}
}

func (server *AuthService) Login(ctx context.Context, req *pd.LoginRequest) (*pd.LoginResponse, error) {
	user, err := server.userStore.Find(req.GetUsername())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot find user: %v", err)
	}

	if user == nil || !user.IsCorrectPassword(req.Password) {
		return nil, status.Errorf(codes.NotFound, "incorrect username or password")
	}

	token, err := server.jwtManager.Generate(user)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot generate access token")
	}

	res := &pd.LoginResponse{
		AccessToken: token,
	}
	return res, nil
}


