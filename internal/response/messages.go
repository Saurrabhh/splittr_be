package response

// System & HTTP Errors
const (
	MsgInternalError = "Something went wrong on our end. Please try again in a moment."
	MsgInvalidBody   = "The information provided is incomplete or invalid. Please check your input."
	MsgUnauthorized  = "Please sign in to continue."
	MsgForbidden     = "You don't have permission to perform this action."
	MsgNotFound      = "We couldn't find the requested resource."
	MsgInvalidParam  = "A required field or parameter is invalid. Please try again."
)

// Validation & Input Errors
const (
	MsgMissingEmailOrPhone   = "Please provide either an email address or phone number."
	MsgInvalidCurrency       = "Currency code must be a valid 3-letter code."
	MsgSelfFriendError       = "You cannot add yourself as a friend."
	MsgNotFriends            = "You are not friends with this user."
	MsgMissingGroupName      = "Please enter a group name."
	MsgInvalidStatusFilter   = "Please select a valid status filter option."
	MsgInvalidAmount         = "Amount must be greater than zero."
	MsgMissingRecipient      = "Please select a recipient for the payment."
	MsgSamePayerPayee        = "Payer and recipient cannot be the same user."
	MsgPayerNotGroupMember   = "The payer must be a member of the group."
	MsgSplitUserNotMember    = "All split users must be members of the group."
	MsgExpenseCreatorOnly    = "Only the person who created this expense can delete it."
	MsgInvalidExpenseFilter  = "Expense filter must be group, personal, or friend."
	MsgAdminRequiredNonActive = "Only admins can query pending or non-active members."
	MsgAlreadyInGroup        = "You are already a member of this group."
	MsgCannotRemoveCreator   = "The group creator cannot be removed."
	MsgAlertRequired         = "Alert payload is required."
)

// Domain Specific Errors
const (
	MsgUserNotFound         = "We couldn't find that user."
	MsgGroupNotFound        = "We couldn't find that group."
	MsgExpenseNotFound      = "We couldn't find that expense."
	MsgNotificationNotFound = "We couldn't find that notification."
	MsgAlreadyGroupMember   = "This user is already a member of the group."
	MsgNotGroupMember       = "You are not a member of this group."
	MsgInvalidSplit         = "The expense split details are invalid."
	MsgInvalidInviteCode    = "The invite code is invalid or has expired."
	MsgAlreadyApproved      = "This membership has already been processed."
	MsgMemberNotFound       = "We couldn't find that member in the group."
	MsgNotGroupAdmin        = "Admin permissions are required for this action."
)

// Mutation Success Messages
const (
	MsgGroupMemberAdded     = "Member added successfully."
	MsgGroupMemberRemoved   = "Member removed successfully."
	MsgNotificationRead     = "Notification marked as read."
	MsgAllNotificationsRead = "All notifications marked as read."
)
