const { SNSClient, PublishCommand } = require("@aws-sdk/client-sns");
const tlsCert = require("./tls-helper");

const notify = new SNSClient();

module.exports.main = async () => {
  if (!process.env.DOMAIN_NAME) {
    throw new Error("DOMAIN_NAME not defined.");
  }
  if (!process.env.SNS_TOPIC_ARN) {
    throw new Error("SNS_TOPIC_ARN not defined.");
  }
  if (!process.env.DAYS_BUFFER) {
    throw new Error("DAYS_BUFFER not defined.");
  }

  const cert = await tlsCert.get(process.env.DOMAIN_NAME);
  const date = new Date(cert.valid_to);
  const now = new Date();
  const daysLeft = Math.floor((date - now) / (1000 * 3600 * 24));

  // Return early if certificate is still good
  if (daysLeft > process.env.DAYS_BUFFER) {
    console.log(
      `${daysLeft} days left till ${process.env.DOMAIN_NAME} expiration > ${process.env.DAYS_BUFFER} days`
    );
    return;
  }

  try {
    const command = new PublishCommand({
      Message: `${process.env.DOMAIN_NAME} certificate will expire in ${daysLeft} days on ${cert.valid_to}.`,
      Subject: `${process.env.DOMAIN_NAME} Certificate Expiring Soon`,
      TopicArn: process.env.SNS_TOPIC_ARN,
    });
    await notify.send(command);
  } catch (err) {
    console.error(err);
  }
}
