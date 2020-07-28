/*
 * Copyright 2020 Amazon.com, Inc. or its affiliates. All Rights Reserved.
 * Licensed under the Apache License, Version 2.0 (the "License").
 * You may not use this file except in compliance with the License.
 * A copy of the License is located at
 *     http://www.apache.org/licenses/LICENSE-2.0
 * or in the "license" file accompanying this file. This file is distributed
 * on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
 * express or implied. See the License for the specific language governing
 * permissions and limitations under the License.
 */

package ad

import (
	"context"
	"encoding/json"
	"errors"
	"esad/internal/controller/es"
	entity "esad/internal/entity/ad"
	"esad/internal/gateway/ad"
	cmapper "esad/internal/mapper"
	mapper "esad/internal/mapper/ad"
	"fmt"
	"github.com/gosuri/uiprogress"
	"io"
	"log"
	"strings"
)

//go:generate mockgen -destination=mocks/mock_ad.go -package=mocks . AnomalyDetectorController

//AnomalyDetectorController is an interface for the AD plugin controllers
type AnomalyDetectorController interface {
	CreateAnomalyDetector(context.Context, entity.CreateDetectorRequest) (*string, error)
	CreateMultiEntityAnomalyDetector(ctx context.Context, request entity.CreateDetectorRequest, interactive bool, display bool) ([]string, error)
	StartDetector(context.Context, string) error
	StopDetector(context.Context, string) error
	DeleteDetector(context.Context, string, bool, bool) error
	DeleteDetectorByName(context.Context, string, bool, bool) error
	StartDetectorByName(context.Context, string, bool) error
	StopDetectorByName(context.Context, string, bool) error
	SearchDetectorByName(context.Context, string) ([]entity.Detector, error)
}

type controller struct {
	reader  io.Reader
	gateway ad.Gateway
	esCtrl  es.Controller
}

func validateCreateRequest(r entity.CreateDetectorRequest) error {
	if len(r.Name) < 1 {
		return fmt.Errorf("name field cannot be empty")
	}
	if len(r.Features) < 1 {
		return fmt.Errorf("features cannot be empty")
	}
	if len(r.Index) < 1 || len(r.Index[0]) < 1 {
		return fmt.Errorf("index field cannot be empty and it should have at least one valid index")
	}
	if len(r.Interval) < 1 {
		return fmt.Errorf("interval field cannot be empty")
	}
	return nil
}
func (c controller) DeleteDetectorByName(ctx context.Context, name string, force bool, display bool) error {
	matchedDetectors, err := c.getDetectorsToProcess(ctx, "delete", name)
	if err != nil {
		return err
	}
	if matchedDetectors == nil {
		return nil
	}
	var bar *uiprogress.Bar
	if display {
		bar = createProgressBar(len(matchedDetectors))
	}
	var failedDetectors []string
	for _, detector := range matchedDetectors {
		if bar != nil {
			bar.Incr()
		}
		err := c.DeleteDetector(ctx, detector.ID, false, force)
		if err != nil {
			failedDetectors = append(failedDetectors, fmt.Sprintf("%s \t Reason: %s", detector.Name, err))
			continue
		}
	}
	if len(failedDetectors) > 0 {
		fmt.Printf("failed to delete %d following detector(s)\n", len(failedDetectors))
		for _, detector := range failedDetectors {
			fmt.Println(detector)
		}
	}
	return nil

}

//NewADController returns new ADController instance
func NewADController(reader io.Reader, esCtrl es.Controller, gateway ad.Gateway) AnomalyDetectorController {
	return &controller{
		reader,
		gateway,
		esCtrl,
	}
}

func (c controller) SearchDetectorByName(ctx context.Context, name string) ([]entity.Detector, error) {
	if len(name) < 1 {
		return nil, fmt.Errorf("detector name cannot be empty")
	}
	payload := entity.SearchRequest{
		Query: entity.SearchQuery{
			Match: entity.Match{
				Name: name,
			},
		},
	}
	response, err := c.gateway.SearchDetector(ctx, payload)
	if err != nil {
		return nil, err
	}
	detectors, err := mapper.MapToDetectors(response, name)
	if err != nil {
		return nil, err
	}
	return detectors, nil
}

func (c controller) StartDetectorByName(ctx context.Context, pattern string, display bool) error {
	return c.processDetectorByAction(ctx, pattern, "start", c.StartDetector, display)
}

func (c controller) getDetectorsToProcess(ctx context.Context, method string, pattern string) ([]entity.Detector, error) {
	if len(pattern) < 1 {
		return nil, fmt.Errorf("name cannot be empty")
	}
	//Search Detector By Name to get ID
	matchedDetectors, err := c.SearchDetectorByName(ctx, pattern)
	if err != nil {
		return nil, err
	}
	if len(matchedDetectors) < 1 {
		fmt.Printf("no detectors matched by name %s\n", pattern)
		return nil, nil
	}
	fmt.Printf("%d detectors matched by name %s\n", len(matchedDetectors), pattern)
	for _, detector := range matchedDetectors {
		fmt.Println(detector.Name)
	}

	proceed := c.askForConfirmation(
		cmapper.StringToStringPtr(
			fmt.Sprintf("esad will %s above matched detector(s). Do you want to proceed? please type (y)es or (n)o and then press enter:", method),
		),
	)
	if !proceed {
		return nil, nil
	}
	return matchedDetectors, nil
}

func (c controller) processDetectorByAction(ctx context.Context, pattern string, action string, f func(c context.Context, s string) error, display bool) error {
	matchedDetectors, err := c.getDetectorsToProcess(ctx, action, pattern)
	if err != nil {
		return err
	}
	if matchedDetectors == nil {
		return nil
	}
	var bar *uiprogress.Bar
	if display {
		bar = createProgressBar(len(matchedDetectors))
	}
	var failedDetectors []string
	for _, detector := range matchedDetectors {
		if bar != nil {
			bar.Incr()
		}
		err := f(ctx, detector.ID)
		if err != nil {
			failedDetectors = append(failedDetectors, fmt.Sprintf("%s \t Reason: %s", detector.Name, err))
			continue
		}
	}
	if len(failedDetectors) > 0 {
		fmt.Printf("\nfailed to %s %d following detector(s)\n", action, len(failedDetectors))
		for _, detector := range failedDetectors {
			fmt.Println(detector)
		}
	}
	return nil
}

func (c controller) StopDetectorByName(ctx context.Context, pattern string, display bool) error {
	return c.processDetectorByAction(ctx, pattern, "stop", c.StopDetector, display)
}

//DeleteDetector deletes detector
func (c controller) DeleteDetector(ctx context.Context, id string, interactive bool, force bool) error {
	if len(id) < 1 {
		return fmt.Errorf("detector Id cannot be empty")
	}
	proceed := true
	if interactive {
		proceed = c.askForConfirmation(
			cmapper.StringToStringPtr(
				fmt.Sprintf(
					"esad will delete detector: %s . Do you want to proceed? please type (y)es or (n)o and then press enter:",
					id,
				),
			),
		)
	}
	if !proceed {
		return nil
	}
	if force {
		res, err := c.gateway.StopDetector(ctx, id) // ignore error
		if err != nil {
			return err
		}
		if interactive {
			fmt.Println(*res)
		}

	}
	err := c.gateway.DeleteDetector(ctx, id)
	if err != nil {
		return err
	}
	return nil
}

func (c controller) askForConfirmation(message *string) bool {

	if message == nil {
		return true
	}
	if len(*message) > 0 {
		fmt.Print(*message)
	}

	var response string
	_, err := fmt.Fscanln(c.reader, &response)
	if err != nil {
		log.Fatal(err)
	}
	switch strings.ToLower(response) {
	case "y", "yes":
		return true
	case "n", "no":
		return false
	default:
		fmt.Printf("I'm sorry but I didn't get what you meant, please type (y)es or (n)o and then press enter:")
		return c.askForConfirmation(cmapper.StringToStringPtr(""))
	}
}

//CreateMultiEntityAnomalyDetector creates multiple detector per entity
func (c controller) CreateMultiEntityAnomalyDetector(ctx context.Context, request entity.CreateDetectorRequest, interactive bool, display bool) ([]string, error) {
	if request.PartitionField == nil || len(*request.PartitionField) < 1 {
		result, err := c.CreateAnomalyDetector(ctx, request)
		if err != nil {
			return nil, err
		}
		return []string{*result}, err
	}
	filterValues, err := getFilterValues(ctx, request, c)
	if err != nil {
		return nil, err
	}
	if len(filterValues) < 1 {
		return nil, fmt.Errorf(
			"failed to get values for partition field: %s, check whether any data is available in index %s",
			*request.PartitionField,
			request.Index,
		)
	}
	proceed := true
	if interactive {
		proceed = c.askForConfirmation(
			cmapper.StringToStringPtr(
				fmt.Sprintf(
					"esad will create %d detector(s). Do you want to proceed? please type (y)es or (n)o and then press enter:",
					len(filterValues),
				),
			),
		)
	}
	if !proceed {
		return nil, nil
	}
	var bar *uiprogress.Bar
	if display {
		bar = createProgressBar(len(filterValues))
	}
	var detectors []string
	name := request.Name
	filter := request.Filter
	var createdDetectors []entity.Detector
	for _, value := range filterValues {
		if bar != nil {
			bar.Incr()
		}
		request.Filter = buildCompoundQuery(*request.PartitionField, value, filter)
		request.Name = fmt.Sprintf("%s-%s", name, value)
		result, err := c.CreateAnomalyDetector(ctx, request)
		if err != nil {
			c.cleanupCreatedDetectors(ctx, createdDetectors)
			return nil, err
		}
		createdDetectors = append(createdDetectors, entity.Detector{
			ID:   *result,
			Name: request.Name,
		})
		detectors = append(detectors, request.Name)
	}
	return detectors, nil
}

func createProgressBar(total int) *uiprogress.Bar {
	if total < 2 {
		return nil
	}
	uiprogress.Start()
	bar := uiprogress.AddBar(total).PrependCompleted()
	bar.Width = 50
	bar.AppendFunc(func(b *uiprogress.Bar) string {
		return fmt.Sprintf("(%d / %d)", b.Current(), total)
	})
	return bar
}

func buildCompoundQuery(field string, value interface{}, userFilter json.RawMessage) json.RawMessage {

	leaf1 := []byte(fmt.Sprintf(`{
    			"bool": {
      				"filter": {
          				"term": {
							"%s" : "%v"
         			 	}
        			}
				}
  			}`, field, value))
	if userFilter == nil {
		return leaf1
	}
	marshal, _ := json.Marshal(entity.Query{
		Bool: entity.Bool{
			Must: []json.RawMessage{
				leaf1, userFilter,
			},
		},
	})
	return marshal
}

func getFilterValues(ctx context.Context, request entity.CreateDetectorRequest, c controller) ([]interface{}, error) {
	var filterValues []interface{}
	for _, index := range request.Index {
		v, err := c.esCtrl.GetDistinctValues(ctx, index, *request.PartitionField)
		if err != nil {
			return nil, err
		}
		filterValues = append(filterValues, v...)
	}
	return filterValues, nil
}

func (c controller) StopDetector(ctx context.Context, ID string) error {
	if len(ID) < 1 {
		return fmt.Errorf("detector Id: %s cannot be empty", ID)
	}
	_, err := c.gateway.StopDetector(ctx, ID)
	if err != nil {
		return err
	}
	return nil
}

func (c controller) StartDetector(ctx context.Context, ID string) error {
	if len(ID) < 1 {
		return fmt.Errorf("detector Id: %s cannot be empty", ID)
	}
	err := c.gateway.StartDetector(ctx, ID)
	if err != nil {
		return err
	}
	return nil
}

func (c controller) CreateAnomalyDetector(ctx context.Context, r entity.CreateDetectorRequest) (*string, error) {

	if err := validateCreateRequest(r); err != nil {
		return nil, err
	}
	payload, err := mapper.MapToCreateDetector(r)
	if err != nil {
		return nil, err
	}
	response, err := c.gateway.CreateDetector(ctx, payload)
	if err != nil {
		return nil, processEntityError(err)
	}
	var data map[string]interface{}
	_ = json.Unmarshal(response, &data)

	detectorID := fmt.Sprintf("%s", data["_id"])
	if !r.Start {
		return cmapper.StringToStringPtr(detectorID), nil
	}

	err = c.StartDetector(ctx, detectorID)
	if err != nil {
		return nil, fmt.Errorf("detector is created with id: %s, but failed to start due to %v", detectorID, err)
	}
	return cmapper.StringToStringPtr(detectorID), nil
}

func processEntityError(err error) error {
	var c entity.CreateError
	data := fmt.Sprintf("%v", err)
	responseErr := json.Unmarshal([]byte(data), &c)
	if responseErr != nil {
		return err
	}
	if len(c.Error.Reason) > 0 {
		return errors.New(c.Error.Reason)
	}
	return err
}

func (c controller) cleanupCreatedDetectors(ctx context.Context, detectors []entity.Detector) {

	if len(detectors) < 1 {
		return
	}
	var deleted []entity.Detector
	for _, d := range detectors {
		err := c.DeleteDetector(ctx, d.ID, false, true)
		if err != nil {
			deleted = append(deleted, d)
		}
	}
	if len(deleted) > 0 {
		var names []string
		for _, d := range deleted {
			names = append(names, d.Name)
		}
		fmt.Println("failed to clean-up created detectors. Please clean up manually following detectors: ", strings.Join(names, ", "))
	}
}