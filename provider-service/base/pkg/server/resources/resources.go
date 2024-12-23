package resources

type HttpHandledResource interface {
	Name() string
	Create(body []byte) (ResponseBody, error)
	Update(id string, body []byte) error
	Delete(id string) error
}

var ResourceTypes = []HttpHandledResource{}

//
//const (
//	Pipeline    = "pipeline"
//	Experiment  = "experiment"
//	RunSchedule = "runschedule"
//	Run         = "run"
//)
//
//func validateType(s string) bool {
//	switch s {
//	case Pipeline, Experiment, RunSchedule, Run:
//		return true
//	default:
//		return false
//	}
//}
//
//func createResource(
//	r *http.Request,
//	resource string,
//) (string, error) {
//	if validateType(resource) {
//		return "foo", nil
//	}
//	return "", errors.New("placeholder")
//}
