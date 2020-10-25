import { APIGatewayProxyHandlerV2 } from "aws-lambda";
import phin = require("phin");

const handler: APIGatewayProxyHandlerV2 = async () => {
  try {
    const result = await phin({
      url: `http://localhost:2772`,
      method: "GET",
      parse: "json"
    });
    return {
      statusCode: 200,
      body: JSON.stringify(result.body)
    };
  } catch (e) {
    return {
      statusCode: 200,
      body: JSON.stringify({ body: e.message })
    };
  }
};

module.exports.handler = handler;
