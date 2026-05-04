package logic

type GetAccountLogic struct{}

func (l *GetAccountLogic) Execute(userID int64) (int64, string) {
    return userID, "demo-user"
}

