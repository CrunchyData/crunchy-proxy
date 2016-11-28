package adapter

type MockAdapter struct{}

func (mc MockAdapter) Do(r *[]byte, i int) error {
	//fmt.Println("Mock request ")
	return nil
}
