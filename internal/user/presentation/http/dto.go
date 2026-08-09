package http

type RegisterRequest struct {
	Name string `json:"name"`
} // @name User.RegisterRequest

type UpdateProfileRequest struct {
	Name            string `json:"name"`
	DefaultCurrency string `json:"defaultCurrency"`
} // @name User.UpdateProfileRequest

type AddFriendRequest struct {
	FriendEmail string `json:"friendEmail"`
	FriendPhone string `json:"friendPhone"`
} // @name User.AddFriendRequest
