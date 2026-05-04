package logic

type HealthLogic struct{}

func (l *HealthLogic) Check() string {
    return "ok"
}

