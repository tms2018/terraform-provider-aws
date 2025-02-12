package macie

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/macie"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_macie_s3_bucket_association")
func ResourceS3BucketAssociation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceS3BucketAssociationCreate,
		ReadWithoutTimeout:   resourceS3BucketAssociationRead,
		UpdateWithoutTimeout: resourceS3BucketAssociationUpdate,
		DeleteWithoutTimeout: resourceS3BucketAssociationDelete,

		Schema: map[string]*schema.Schema{
			"bucket_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"prefix": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"member_account_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidAccountID,
			},
			"classification_type": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"continuous": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      macie.S3ContinuousClassificationTypeFull,
							ValidateFunc: validation.StringInSlice([]string{macie.S3ContinuousClassificationTypeFull}, false),
						},
						"one_time": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      macie.S3OneTimeClassificationTypeNone,
							ValidateFunc: validation.StringInSlice([]string{macie.S3OneTimeClassificationTypeFull, macie.S3OneTimeClassificationTypeNone}, false),
						},
					},
				},
			},
		},
	}
}

func resourceS3BucketAssociationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).MacieConn()

	req := &macie.AssociateS3ResourcesInput{
		S3Resources: []*macie.S3ResourceClassification{
			{
				BucketName:         aws.String(d.Get("bucket_name").(string)),
				ClassificationType: expandClassificationType(d),
			},
		},
	}
	if v, ok := d.GetOk("member_account_id"); ok {
		req.MemberAccountId = aws.String(v.(string))
	}
	if v, ok := d.GetOk("prefix"); ok {
		req.S3Resources[0].Prefix = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating Macie S3 bucket association: %#v", req)
	resp, err := conn.AssociateS3ResourcesWithContext(ctx, req)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Macie S3 bucket association: %s", err)
	}
	if len(resp.FailedS3Resources) > 0 {
		return sdkdiag.AppendErrorf(diags, "creating Macie S3 bucket association: %s", resp.FailedS3Resources[0])
	}

	d.SetId(fmt.Sprintf("%s/%s", d.Get("bucket_name"), d.Get("prefix")))
	return append(diags, resourceS3BucketAssociationRead(ctx, d, meta)...)
}

func resourceS3BucketAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).MacieConn()

	req := &macie.ListS3ResourcesInput{}
	if v, ok := d.GetOk("member_account_id"); ok {
		req.MemberAccountId = aws.String(v.(string))
	}

	bucketName := d.Get("bucket_name").(string)
	prefix := d.Get("prefix")

	var res *macie.S3ResourceClassification
	err := conn.ListS3ResourcesPagesWithContext(ctx, req, func(page *macie.ListS3ResourcesOutput, lastPage bool) bool {
		for _, v := range page.S3Resources {
			if aws.StringValue(v.BucketName) == bucketName && aws.StringValue(v.Prefix) == prefix {
				res = v
				return false
			}
		}

		return true
	})
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing Macie S3 bucket associations: %s", err)
	}

	if res == nil {
		log.Printf("[WARN] Macie S3 bucket association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err := d.Set("classification_type", flattenClassificationType(res.ClassificationType)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting classification_type: %s", err)
	}

	return diags
}

func resourceS3BucketAssociationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).MacieConn()

	if d.HasChange("classification_type") {
		req := &macie.UpdateS3ResourcesInput{
			S3ResourcesUpdate: []*macie.S3ResourceClassificationUpdate{
				{
					BucketName:               aws.String(d.Get("bucket_name").(string)),
					ClassificationTypeUpdate: expandClassificationTypeUpdate(d),
				},
			},
		}
		if v, ok := d.GetOk("member_account_id"); ok {
			req.MemberAccountId = aws.String(v.(string))
		}
		if v, ok := d.GetOk("prefix"); ok {
			req.S3ResourcesUpdate[0].Prefix = aws.String(v.(string))
		}

		log.Printf("[DEBUG] Updating Macie S3 bucket association: %#v", req)
		resp, err := conn.UpdateS3ResourcesWithContext(ctx, req)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Macie S3 bucket association: %s", err)
		}
		if len(resp.FailedS3Resources) > 0 {
			return sdkdiag.AppendErrorf(diags, "updating Macie S3 bucket association: %s", resp.FailedS3Resources[0])
		}
	}

	return append(diags, resourceS3BucketAssociationRead(ctx, d, meta)...)
}

func resourceS3BucketAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).MacieConn()

	log.Printf("[DEBUG] Deleting Macie S3 bucket association: %s", d.Id())

	req := &macie.DisassociateS3ResourcesInput{
		AssociatedS3Resources: []*macie.S3Resource{
			{
				BucketName: aws.String(d.Get("bucket_name").(string)),
			},
		},
	}
	if v, ok := d.GetOk("member_account_id"); ok {
		req.MemberAccountId = aws.String(v.(string))
	}
	if v, ok := d.GetOk("prefix"); ok {
		req.AssociatedS3Resources[0].Prefix = aws.String(v.(string))
	}

	resp, err := conn.DisassociateS3ResourcesWithContext(ctx, req)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Macie S3 bucket association: %s", err)
	}
	if len(resp.FailedS3Resources) > 0 {
		failed := resp.FailedS3Resources[0]
		// {
		// 	ErrorCode: "InvalidInputException",
		// 	ErrorMessage: "The request was rejected. The specified S3 resource (bucket or prefix) is not associated with Macie.",
		// 	FailedItem: {
		// 	 BucketName: "tf-macie-example-002"
		// 	}
		// }
		if aws.StringValue(failed.ErrorCode) == macie.ErrCodeInvalidInputException &&
			strings.Contains(aws.StringValue(failed.ErrorMessage), "is not associated with Macie") {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting Macie S3 bucket association: %s", failed)
	}

	return diags
}
