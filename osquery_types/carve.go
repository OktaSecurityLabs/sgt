package osquery_types

import (
	"encoding/base64"
	"fmt"
	"os"
)

type FileCarve struct {
	SessionID string
	Chunks    []*CarveData
}

func (fc FileCarve) RebuildCarve() ([]byte, error) {
	data := []byte{}
	for _, d := range fc.Chunks {
		decoded, err := base64.StdEncoding.DecodeString(d.Data)
		if err != nil {
			return nil, err
		}
		data = append(data, decoded...)
	}
	return data, nil
}

func (fc FileCarve) SaveToFile(path string) error {
	data, err := fc.RebuildCarve()
	if err != nil {
		return err
	}
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	file.Write(data)
	return nil
}

type Carve struct {
	// {"block_count":"1","block_size":"300000","carve_size":"12800","carve_id":"3bed2f21-e306-4d2b-962c-3b207946c298","request_id":"","node_key":"hx9gvmkir0xcta0opsb5"}
	BlockCount string `json:"block_count"`
	BlockSize  string `json:"block_size"`
	CarveSize  string `json:"carve_size"`
	CarveID    string `json:"carve_id"`
	RequestID  string `json:"request_id"`
	NodeKey    string `json:"node_key"`
	SessionID  string `json:"session_id"`
}

type CarveData struct {
	// {"block_id":"0","session_id":"306959833118746","request_id":"14998","data":"Q2IgUmVZJHGaVuFAKdx\/uNio7D2lkghvy6P42HvcpRFagBTKkBsg8i4wii8.......AAiGArAEHcDDL+4pydnAq7FxuL3DmGTBFcscAXhPGhx"}
	//{"block_id":"2","session_id":"fWWSLRmqFbMxotP","request_id":"","data":"G92GKjia814OBIl7uQaKfz1Qk70FoOOuk7DBKQhbS86QC2yypy26fa2u3khh1+V0zftjTgTtH13z\/khv8TJ0eKOju5hYGPMSBOF4KCyIYA4dd520Z4u9lwDjX4hvpuQ+hVkSS1wBKtu+Rdf8r+dHzAHgB3MrwJdYclc3wZSv4xmXWi63rhRuOZJyMpF\/IfB8\/ae7zwBXrW9DFU93+1ud795W+S5s2vStHDPMr8eqBf+A6VvplgN\/\/+1HC0IkvDSgiqdz6Q\/b6j\/eXQBMfApKzHQpTPKCpT8PX2XX7RxeWJYjMcPD4YVlVHt\/5PIDJYkn84K1oORM2kqn1b+9uwBYeTrGFt26kdfoF+AFYMwzVPuuZfncOCdCho\/lc2Pi+FqMecYphTn1qUh0AaOLUJpb\/1FNAJbcNgkZfNw1zisPnwG\/CKjt6O5l\/GDR5pwIGTnccsEWdPcyhNruni2nvjVJPmTpfZwvrj0sfbsmAKZlNtjjcyEOkkOdwElruffbrFrwTG60jDRWLXiGSvl6d0Qd5HukLO1H8Aof7\/+3\/wPrz0Tatv6EjFlD4NZ9rTfSJ1flQ8QookAXRm90+kBOS4GOQNg2bGFm+pYTgIVrAzzveISfGA0yRnq4g61SDa\/ntnO69\/udsYabzummEl6PtVW3K8hhFjAaRAE873gWrg0gFYDDoqMRamZuZl8pnZk3rqxnzXlDDnNqeKw57yfocH1\/9tCsIUh9GGZyWHQ0pAIg3no\/BB\/C5mD5s9Yx3+qQ0Ixo8oOGRGxuwuoQqXLQBZI4AoIPOZ6nAuD92XRkNch8+u\/35lEQVR+h5I1Y4oOGxfa+p4jCR5DKPXvWQmA0iGpA8KH\/Cf0zQOuxKH+AYgsZQChnq5bcwU3nVLMnIGPcuSgk5o7+KiVZwxrnS6FLMwAkn19eRKsZ2VMCSAHKg1i\/QczGXGjIA5KNxNHrKM\/1QeYQgJrBshtKEm3bkPqD+SRqTvLlCvEoq8\/7fQ4E5IPV5\/0ey6P9xSWyholA6un0lNskYs5ERPXI7K1\/FlBOK43KT7h14FCBsMS9T7o+V2S+JTQGCI8gnD1B4nVPxfqttTCpDKEU6HAbwryS7Y0bAEK8gq5uQ2WtB6ShcH4rXvdUiZJTEKrgtggZ0yF9kPJXBMWGKaGSGaz\/R6R6Beln3+8IELKIklMkQTAJJWUu9mkpwNrXuOnsMVdvb9hYfd6fsOb1XJRAt\/0W+N5hkti2uu1IxjYAIVwFTxP94dBa\/1MISxj\/IZ+jd+22oDGtEiVa8hn9EqIyeMWxf+Q7VMTFzZg+1xd5wFctEsyHicPsA3+EdH7+Ue\/Bd\/AzWBR6uzEq24DbNHZFh2D1hyRSfhRItgZZSUAibcKrEpsGjTrNANaUwatmG1EkalXZpPyoBK\/ZrUFZ2wEEIEIkY9fla7iwxCBycBlOo5q9Ztk48WzjyBZuFpYQ9TolMGtFxAI2wOwnRP1ghsADG2Q\/ANNTSN0rMfolZ5PP0m05mXVsXMCTg85oddBByBLEhWw18PQIXoKJX5Kg\/i9ekP1KkMas+c1t+298kKLa3IbU2R7Dp7KmAhDqVYm2u3JxUjQmSWNSOSL7mzcIZDgN2ZSfG36kd0k80YPVjFb9qH3CWvB8kP77GiQBQcawgpJ6n3OHy1oHUC7HkUePJAy3o43JZxawIMSxXPXQ5OxvnjPa1x2OFcfmEobngn8sUbxNos1WbJLpKFNLIM4xwZg\/J6wcleGdGwMi+h9Ycxwmyr7fXeh9BW22SuK2dxBRT2KYyZASXNSqCg7DyuOyvXEDwNrjUMFh6KzjMBI+i6iHuO0dif3PHdD0dvYHEgLQ7kjY9085tPQAK\/BKc9xRcA75hKQEgs0Ez+6UTFY9aP0b55+WNZIsJFacTvu69+dAQD5oX\/d+hDzdZWDJ4yTWB8ObtJa6JSuuLSOqr+fjnGjdMuD5x+Bx2v6\/cJDARKfhFWegNfnkYrCAfJ0V15aTuADvNeKc8gDbOEnvysVctb6QPQEZw8VhXozVSZROxhAS4hhU7+vQfwCg38TGLlwpU2LSo0kNfvFMypVPZktADphU+CR+6cxkB5S9N5BUYAsh4au\/hlQAor7fg\/cqwiPzNSnNXiFkQBAszfbmOUCqqxAyyCfRtHVJtQlfxR71e0gFYJv5LVZtIq80Nmkadq94Dpev+0wOFGSDy9d9hqB4jlN88wjDJ41H3MQ2\/7eQCsCdi0Li+GVsRObLACQOoiGgChQL1+2ZyeqgwKU\/bKNYuA5UwSXhyEHplgpsFeL4Ze5cFEK9E4CobsLI7txSmllc6nUp5+K3LMmHiFFE0LwYKefWai7lAOWDFd2Ivk3pWzUBiKvPgXg5F8Ig0QVC1zktLddx7fqTc6NlpHHt+pNpbv0K1rpnzDP\/ohH\/RRw\/l\/5bE4BbF23DVJ\/rj9nPBTKpeSMmUWYFn1879o+KL7lvGmVWIMUkogrZe14lkDLxwYif49ZF2\/rf3r2VfQwpKvkcUfbTAHEVvOBkim03sPyxsesytvwxj4mFG\/AKJ9eKbuSANAuLVBUoP1b\/0e4CUDEb0XZTPjHrKYSTVBtDoXAJ71RX5EjM8PBOdQVB8a+wUeL1k2O\/CgHabKK7vFseht0F4PsX9KArj2JkUoApR29hHbn10m9aypV3fjM\/QoaIpXf9E17TUkyV3FLvQZKBTYGVoCuPcttf7xaIs\/f0KqO70fKLKH+KKxaVo9RGVfAF+Ed8nY6ftFHY+RVWLGrsQJLPP1akKfw2qng1puKeIdeCUkmxDSPfRUaDKBjRueB5ZPQkNq2ymWOueyGdPhD3gV+4mmrLGr5wX+M6j7TfdSTN1dX4wdXoiqNdSHKtuSA9d9ZiwyfpXPD8ni0GVknjnh+6XHayAWJGhItjiyuggktpKv4HHQ98Km+q9sLS+z5J0HYXwvsb4qqjOe9yK5YkBlMAvT8cqMnAGrbUD2OiZ1H+7HxyB+2JpOgBBlTwCSg9wJIfX49WXaw+a2eupLU\/PAFl2hHqq4hGKxpFmn73WYh\/uq8mA6Nj\/eWo0mp0X77K4G5InNqDovtt9Aaiync54hfrWb48W9\/q5cslm086B19dg\/TPACAs12hsBAiBczuvttP52QGLR+6b0sV3T0WWHsTzZzkDRqM8WFLqXflpJfA+jH4IQSdh+EwmhSOD4GS07cD3G7twpF+AWL+IMmfvq3Dke1O75N7LUC1rXJ3cBljT6pGWjpVBesJVRkcbEOIuwuhxbjn\/zRG932X3TyfwT8WaBSj\/DIQsYQyYBi0dK4NEia600zVvCKVjYWwUj05LzAjlkk6iQIevYdkI+smhFY\/GzYBh75E0H\/4RjJoDnFYrHp148+Rt3HkvpMWjg77z36u45v6pH5Pl4xNhcE6Xf8LyK4x5AyF\/hzXvEFV3IsXu5eONLeEXJiDkVKz5AEq68vGWyXjKuVGZhPGNXDXcJiF30oso71g0vPLx7oqCpY\/\/CGn\/kqiXxtEFBkCqrApBfy5eKZPpkGQ7mfg8CFmmlhlLYU0Jo124mkoitk1Yc1mzdvfrNyTStb8ZYnUHnXMu2V8CrkEMZ2ER5nqsfhu\/2EA7ggGQMh4cnSZyxphqD1R6nNNJzeGlBLQkLxeiLpVrU+lx34mr7hp7ClajwlpXc8nEb0Pl+sFkXxvcfH7T6S8Rl\/\/FrbMNUBJ10EijYOoYZ2LQsXM+2e0VJ0sGde0bRasfDJKqK8IDXfkXOj\/9y8F8a\/AL+rvlVehdj6MKSeeMFSEYACL9UfcaK3weEMmOSBUg7tmIfqtrsN88sMde8uSJqB3rkU1TidLt9pjuuYMAyUD0S2DK7xBNPZfvzX5hsN8+MJV+1ZwXibuXg3TS1sj6wKECax0vEBB3Lz8Q5sOQhq8VdKxfhdd8BbqvcezehyQsCN8V2Y56VrNq3uLBKH71GMKmXlhE9WvY6Cn85hoh48gYSZ8HJbDmKbT42oEyH4bqoXjzhX+isn0pxv5\/\/NYksntcCLJDut9v"}
	SessionBlockID string `json:"session_block_id"`
	BlockID        string `json:"block_id"`
	SessionID      string `json:"session_id"`
	RequestID      string `json:"request_id"`
	Data           string `json:"data"`
	TimeToLive     int64  `json:"time_to_live"`
}

func (cd CarveData) SetSBID() string {
	return fmt.Sprintf("%s-%s", cd.SessionID, cd.BlockID)
}
