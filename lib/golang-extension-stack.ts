import * as cdk from "@aws-cdk/core";
import * as lambda from "@aws-cdk/aws-lambda";
import * as apigwv2 from "@aws-cdk/aws-apigatewayv2";
import { fromRoot, getFunctionPath } from "./utils/utils";
import * as ssm from "@aws-cdk/aws-ssm";

export class GolangExtension extends cdk.Stack {
  constructor(scope: cdk.Construct, id: string, props?: cdk.StackProps) {
    super(scope, id, props);

    const ssmIPCLayer = new lambda.LayerVersion(this, "ssmIPCLayer", {
      code: lambda.Code.fromAsset(fromRoot("./layers"))
    });

    const helloHandler = new lambda.Function(this, "helloHandler", {
      code: lambda.Code.fromAsset(getFunctionPath("hello")),
      handler: "handler.handler",
      runtime: lambda.Runtime.NODEJS_12_X
    });
    helloHandler.addLayers(ssmIPCLayer);

    const ssmParam = new ssm.StringParameter(this, "ssmIPCParam", {
      parameterName: "extension-parameter",
      stringValue: "Hey im from SSM!"
    });

    ssmParam.grantRead(helloHandler);

    const api = new apigwv2.HttpApi(this, "httpApi");

    const helloIntegration = new apigwv2.LambdaProxyIntegration({
      handler: helloHandler
    });

    api.addRoutes({
      path: "/",
      methods: [apigwv2.HttpMethod.GET],
      integration: helloIntegration
    });

    new cdk.CfnOutput(this, "apiUrl", {
      value: api.url ?? "error"
    });
  }
}
