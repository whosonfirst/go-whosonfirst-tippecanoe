// Code generated by smithy-go-codegen DO NOT EDIT.

package iam

import (
	"context"
	"fmt"
	awsmiddleware "github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/aws/smithy-go/middleware"
	smithyhttp "github.com/aws/smithy-go/transport/http"
	"time"
)

// After you generate a group or policy report using the
// GenerateServiceLastAccessedDetails operation, you can use the JobId parameter
// in GetServiceLastAccessedDetailsWithEntities . This operation retrieves the
// status of your report job and a list of entities that could have used group or
// policy permissions to access the specified service.
//   - Group – For a group report, this operation returns a list of users in the
//     group that could have used the group’s policies in an attempt to access the
//     service.
//   - Policy – For a policy report, this operation returns a list of entities
//     (users or roles) that could have used the policy in an attempt to access the
//     service.
//
// You can also use this operation for user or role reports to retrieve details
// about those entities. If the operation fails, the
// GetServiceLastAccessedDetailsWithEntities operation returns the reason that it
// failed. By default, the list of associated entities is sorted by date, with the
// most recent access listed first.
func (c *Client) GetServiceLastAccessedDetailsWithEntities(ctx context.Context, params *GetServiceLastAccessedDetailsWithEntitiesInput, optFns ...func(*Options)) (*GetServiceLastAccessedDetailsWithEntitiesOutput, error) {
	if params == nil {
		params = &GetServiceLastAccessedDetailsWithEntitiesInput{}
	}

	result, metadata, err := c.invokeOperation(ctx, "GetServiceLastAccessedDetailsWithEntities", params, optFns, c.addOperationGetServiceLastAccessedDetailsWithEntitiesMiddlewares)
	if err != nil {
		return nil, err
	}

	out := result.(*GetServiceLastAccessedDetailsWithEntitiesOutput)
	out.ResultMetadata = metadata
	return out, nil
}

type GetServiceLastAccessedDetailsWithEntitiesInput struct {

	// The ID of the request generated by the GenerateServiceLastAccessedDetails
	// operation.
	//
	// This member is required.
	JobId *string

	// The service namespace for an Amazon Web Services service. Provide the service
	// namespace to learn when the IAM entity last attempted to access the specified
	// service. To learn the service namespace for a service, see Actions, resources,
	// and condition keys for Amazon Web Services services (https://docs.aws.amazon.com/service-authorization/latest/reference/reference_policies_actions-resources-contextkeys.html)
	// in the IAM User Guide. Choose the name of the service to view details for that
	// service. In the first paragraph, find the service prefix. For example, (service
	// prefix: a4b) . For more information about service namespaces, see Amazon Web
	// Services service namespaces (https://docs.aws.amazon.com/general/latest/gr/aws-arns-and-namespaces.html#genref-aws-service-namespaces)
	// in the Amazon Web Services General Reference.
	//
	// This member is required.
	ServiceNamespace *string

	// Use this parameter only when paginating results and only after you receive a
	// response indicating that the results are truncated. Set it to the value of the
	// Marker element in the response that you received to indicate where the next call
	// should start.
	Marker *string

	// Use this only when paginating results to indicate the maximum number of items
	// you want in the response. If additional items exist beyond the maximum you
	// specify, the IsTruncated response element is true . If you do not include this
	// parameter, the number of items defaults to 100. Note that IAM might return fewer
	// results, even when there are more results available. In that case, the
	// IsTruncated response element returns true , and Marker contains a value to
	// include in the subsequent call that tells the service where to continue from.
	MaxItems *int32

	noSmithyDocumentSerde
}

type GetServiceLastAccessedDetailsWithEntitiesOutput struct {

	// An EntityDetailsList object that contains details about when an IAM entity
	// (user or role) used group or policy permissions in an attempt to access the
	// specified Amazon Web Services service.
	//
	// This member is required.
	EntityDetailsList []types.EntityDetails

	// The date and time, in ISO 8601 date-time format (http://www.iso.org/iso/iso8601)
	// , when the generated report job was completed or failed. This field is null if
	// the job is still in progress, as indicated by a job status value of IN_PROGRESS .
	//
	// This member is required.
	JobCompletionDate *time.Time

	// The date and time, in ISO 8601 date-time format (http://www.iso.org/iso/iso8601)
	// , when the report job was created.
	//
	// This member is required.
	JobCreationDate *time.Time

	// The status of the job.
	//
	// This member is required.
	JobStatus types.JobStatusType

	// An object that contains details about the reason the operation failed.
	Error *types.ErrorDetails

	// A flag that indicates whether there are more items to return. If your results
	// were truncated, you can make a subsequent pagination request using the Marker
	// request parameter to retrieve more items. Note that IAM might return fewer than
	// the MaxItems number of results even when there are more results available. We
	// recommend that you check IsTruncated after every call to ensure that you
	// receive all your results.
	IsTruncated bool

	// When IsTruncated is true , this element is present and contains the value to use
	// for the Marker parameter in a subsequent pagination request.
	Marker *string

	// Metadata pertaining to the operation's result.
	ResultMetadata middleware.Metadata

	noSmithyDocumentSerde
}

func (c *Client) addOperationGetServiceLastAccessedDetailsWithEntitiesMiddlewares(stack *middleware.Stack, options Options) (err error) {
	if err := stack.Serialize.Add(&setOperationInputMiddleware{}, middleware.After); err != nil {
		return err
	}
	err = stack.Serialize.Add(&awsAwsquery_serializeOpGetServiceLastAccessedDetailsWithEntities{}, middleware.After)
	if err != nil {
		return err
	}
	err = stack.Deserialize.Add(&awsAwsquery_deserializeOpGetServiceLastAccessedDetailsWithEntities{}, middleware.After)
	if err != nil {
		return err
	}
	if err := addProtocolFinalizerMiddlewares(stack, options, "GetServiceLastAccessedDetailsWithEntities"); err != nil {
		return fmt.Errorf("add protocol finalizers: %v", err)
	}

	if err = addlegacyEndpointContextSetter(stack, options); err != nil {
		return err
	}
	if err = addSetLoggerMiddleware(stack, options); err != nil {
		return err
	}
	if err = awsmiddleware.AddClientRequestIDMiddleware(stack); err != nil {
		return err
	}
	if err = smithyhttp.AddComputeContentLengthMiddleware(stack); err != nil {
		return err
	}
	if err = addResolveEndpointMiddleware(stack, options); err != nil {
		return err
	}
	if err = v4.AddComputePayloadSHA256Middleware(stack); err != nil {
		return err
	}
	if err = addRetryMiddlewares(stack, options); err != nil {
		return err
	}
	if err = awsmiddleware.AddRawResponseToMetadata(stack); err != nil {
		return err
	}
	if err = awsmiddleware.AddRecordResponseTiming(stack); err != nil {
		return err
	}
	if err = addClientUserAgent(stack, options); err != nil {
		return err
	}
	if err = smithyhttp.AddErrorCloseResponseBodyMiddleware(stack); err != nil {
		return err
	}
	if err = smithyhttp.AddCloseResponseBodyMiddleware(stack); err != nil {
		return err
	}
	if err = addSetLegacyContextSigningOptionsMiddleware(stack); err != nil {
		return err
	}
	if err = addOpGetServiceLastAccessedDetailsWithEntitiesValidationMiddleware(stack); err != nil {
		return err
	}
	if err = stack.Initialize.Add(newServiceMetadataMiddleware_opGetServiceLastAccessedDetailsWithEntities(options.Region), middleware.Before); err != nil {
		return err
	}
	if err = awsmiddleware.AddRecursionDetection(stack); err != nil {
		return err
	}
	if err = addRequestIDRetrieverMiddleware(stack); err != nil {
		return err
	}
	if err = addResponseErrorMiddleware(stack); err != nil {
		return err
	}
	if err = addRequestResponseLogging(stack, options); err != nil {
		return err
	}
	if err = addDisableHTTPSMiddleware(stack, options); err != nil {
		return err
	}
	return nil
}

func newServiceMetadataMiddleware_opGetServiceLastAccessedDetailsWithEntities(region string) *awsmiddleware.RegisterServiceMetadata {
	return &awsmiddleware.RegisterServiceMetadata{
		Region:        region,
		ServiceID:     ServiceID,
		OperationName: "GetServiceLastAccessedDetailsWithEntities",
	}
}
