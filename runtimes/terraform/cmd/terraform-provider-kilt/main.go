package main

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/sysdiglabs/agent-kilt/runtimes/terraform/tf"
	"log"
)

func main() {
	err := tfsdk.Serve(context.Background(), tf.New, tfsdk.ServeOpts{
		Name: "kilt",
	})
	if err != nil {
		log.Fatalf("terraform-provider-kilt plugin failed: %+v", err)
	}
}
