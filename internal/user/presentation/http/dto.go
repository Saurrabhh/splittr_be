package http

type RegisterRequest struct {
	Name string `json:"name"`
} // @name RegisterRequest

type UpdateProfileRequest struct {
	Name            string `json:"name"`
	DefaultCurrency string `json:"defaultCurrency"`
} // @name UpdateProfileRequest

type AddFriendRequest struct {
	FriendEmail string `json:"friendEmail"`
	FriendPhone string `json:"friendPhone"`
} // @name AddFriendRequest
