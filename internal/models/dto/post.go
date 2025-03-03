package dto

type Pagination struct {
	Offset int `form:"offset" validate:"min=0"`
	Limit  int `form:"limit"  validate:"min=1,max=20"`
}

type UserFeedRequest struct {
	Search     *string   `form:"search"`
	Tags       *[]string `form:"tags"`
	Pagination `          form:"pagination"`
	ID         int64 `form:"-"`
}
