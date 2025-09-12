package provider

import (
	"context"
	"fmt"
	"github.com/kubeflow/pipelines/backend/api/v2beta1/go_client"
	"github.com/samber/lo"
	"github.com/sky-uk/kfp-operator/pkg/providers/base"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/util"
	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/client"
	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/client/resource"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gopkg.in/yaml.v2"
)

type RecurringRunService interface {
	CreateRecurringRun(
		ctx context.Context,
		rsd base.RunScheduleDefinition,
		pipelineId string,
		pipelineVersionId string,
		experimentId string,
	) (string, error)
	GetRecurringRun(ctx context.Context, id string) (string, error)
	DeleteRecurringRun(ctx context.Context, id string) error
}

type DefaultRecurringRunService struct {
	client         client.RecurringRunServiceClient
	labelGenerator LabelGen
}

func NewRecurringRunService(conn *grpc.ClientConn, labelGenerator LabelGen) (RecurringRunService, error) {
	if conn == nil {
		return nil, fmt.Errorf(
			"no gRPC connection was provided to start recurring run service",
		)
	}

	return &DefaultRecurringRunService{
		client:         go_client.NewRecurringRunServiceClient(conn),
		labelGenerator: labelGenerator,
	}, nil
}

// CreateRecurringRun creates a recurring run and returns the recurring run id.
func (rrs *DefaultRecurringRunService) CreateRecurringRun(
	ctx context.Context,
	rsd base.RunScheduleDefinition,
	pipelineId string,
	pipelineVersionId string,
	experimentId string,
) (string, error) {
	// needed to write metadata of the recurring run as no other field is possible
	runScheduleAsDescription, err := yaml.Marshal(resource.References{
		PipelineName:         rsd.PipelineName,
		RunConfigurationName: rsd.RunConfigurationName,
		Artifacts:            rsd.Artifacts,
	})
	if err != nil {
		return "", err
	}

	recurringRunName, err := util.ResourceNameFromNamespacedName(rsd.Name)
	if err != nil {
		return "", err
	}

	generatedLabels, err := rrs.labelGenerator.GenerateLabels(rsd)
	if err != nil {
		return "", err
	}

	runtimeParams := make(map[string]*structpb.Value)
	for k, v := range lo.Assign(rsd.Parameters, generatedLabels) {
		runtimeParams[k] = structpb.NewStringValue(v)
	}

	apiCronSchedule, err := createAPICronSchedule(rsd)
	if err != nil {
		return "", err
	}

	recurringRun, err := rrs.client.CreateRecurringRun(ctx, &go_client.CreateRecurringRunRequest{
		RecurringRun: &go_client.RecurringRun{
			DisplayName: recurringRunName,
			Description: string(runScheduleAsDescription),
			PipelineSource: &go_client.RecurringRun_PipelineVersionReference{
				PipelineVersionReference: &go_client.PipelineVersionReference{
					PipelineId:        pipelineId,
					PipelineVersionId: pipelineVersionId,
				},
			},
			RuntimeConfig: &go_client.RuntimeConfig{
				Parameters: runtimeParams,
			},
			MaxConcurrency: 1,
			Trigger: &go_client.Trigger{
				Trigger: &go_client.Trigger_CronSchedule{CronSchedule: apiCronSchedule},
			},
			Mode:         go_client.RecurringRun_ENABLE,
			NoCatchup:    true,
			ExperimentId: experimentId,
		},
	})
	if err != nil {
		return "", err
	}

	return recurringRun.RecurringRunId, nil
}

// GetRecurringRun takes a recurring run id and returns the recurring run description.
func (rrs *DefaultRecurringRunService) GetRecurringRun(
	ctx context.Context,
	id string,
) (string, error) {
	recurringRun, err := rrs.client.GetRecurringRun(
		ctx,
		&go_client.GetRecurringRunRequest{RecurringRunId: id},
	)
	if err != nil {
		return "", err
	}

	return recurringRun.Description, nil
}

// DeleteRecurringRun deletes a recurring run by recurring run id.
// Does not error if there is no such recurring run id.
func (rrs *DefaultRecurringRunService) DeleteRecurringRun(
	ctx context.Context,
	id string,
) error {
	if _, err := rrs.client.DeleteRecurringRun(
		ctx,
		&go_client.DeleteRecurringRunRequest{RecurringRunId: id},
	); err != nil {
		if st, ok := status.FromError(err); ok {
			// is a gRPC error
			switch st.Code() {
			case codes.NotFound:
				return nil
			default:
				return err
			}
		}
		return err
	}

	return nil
}

func createAPICronSchedule(
	rsd base.RunScheduleDefinition,
) (*go_client.CronSchedule, error) {
	cronExpression, err := util.ParseCron(rsd.Schedule.CronExpression)
	if err != nil {
		return nil, err
	}

	schedule := &go_client.CronSchedule{
		Cron: cronExpression.PrintGo(),
	}

	if rsd.Schedule.StartTime != nil {
		schedule.StartTime = timestamppb.New(rsd.Schedule.StartTime.Time)
	}

	if rsd.Schedule.EndTime != nil {
		schedule.EndTime = timestamppb.New(rsd.Schedule.EndTime.Time)
	}

	return schedule, nil
}
