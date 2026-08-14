package http

import "github.com/Saurrabhh/splittr_be/internal/user/domain"

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

type AddFriendResponse struct {
	Friend domain.User `json:"friend"`
	Status string      `json:"status"`
} // @name User.AddFriendResponse

type UpdateFriendStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=ACCEPTED DECLINED BLOCKED"`
} // @name User.UpdateFriendStatusRequest

type UserSettingsResponse struct {
	AutoAcceptFriendRequests bool `json:"autoAcceptFriendRequests"`
} // @name User.UserSettingsResponse

type UpdateUserSettingsRequest struct {
	AutoAcceptFriendRequests bool `json:"autoAcceptFriendRequests"`
} // @name User.UpdateUserSettingsRequest
