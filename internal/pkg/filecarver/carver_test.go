package filecarver

//{"block_count":"21","block_size":"300000","carve_size":"6261760","carve_id":"b5f73c17-6e0c-4c49-9ad3-70c25aba41a9","request_id":"14998","node_key":"b3NxdWVyeTo3ZAGMyMGNiNS02ZAjdmLTc5NGUtNTk4NS1iYmViMmJjZAmNkZAjYZD"}
//var startTest = osquery_types.Carve {
	//BlockCount: 21,
	//BlockSize: 300000,
	//CarveSize: 6261760,
	//CarveID:  "b5f73c17-6e0c-4c49-9ad3-70c25aba41a9",
//}

//var dataTest = osquery_types.CarveData{
	//"",
	//"",
	//"",
	//"",
//}


/*func TestStartCarve(t *testing.T) {
	mockdb := helpers.NewMockDB()

	handler := StartCarve(mockdb)

	test := helpers.GenerateHandleTester(t, handler)

	js, err := json.Marshal(startTest)
	if err != nil {
		t.Error(err)
	}

	w := test("POST", "/carve/start", url.Values{}, bytes.NewReader(js))

	if w.Code != http.StatusOK {
		t.Errorf("returned status: %d", w.Code)
	}

	body, err := ioutil.ReadAll(w.Body)
	if err != nil {
		t.Error(err)
	}
	type statusSuccess struct {
		Success bool `json:"success"`
		SessionID string `json:"session_id"`
	}
	ok := statusSuccess{}
	err = json.Unmarshal(body, &ok)
	if err != nil {
		t.Error(err)
	}
	if !ok.Success {
		t.Errorf("%+v", ok)
	}
}
*/