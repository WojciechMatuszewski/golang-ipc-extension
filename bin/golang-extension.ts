#!/usr/bin/env node
import "source-map-support/register";
import * as cdk from "@aws-cdk/core";
import { GolangExtension } from "../lib/golang-extension-stack";

const app = new cdk.App();
new GolangExtension(app, "GolangExtensionStack", {
  env: { region: "eu-central-1" }
});
