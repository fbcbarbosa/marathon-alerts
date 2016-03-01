# marathon-alerts

Marathon Alerts is a tool for monitoring the apps running on marathon. Inspired from [kubernetes-alerts](https://github.com/AcalephStorage/kubernetes-alerts) and [consul-alerts](https://github.com/AcalephStorage/consul-alerts).

This was initially built for Marathon 0.8.0, hence we don't use the event bus.

## Usage
```
$ marathon-alerts --help
Usage of marathon-alerts:
      --alerts-suppress-duration duration           Suppress alerts for this duration once notified (default 30m0s)
      --check-interval duration                     Check runs periodically on this interval (default 30s)
      --check-min-healthy-fail-threshold value      Min instances check fail threshold (default 0.6)
      --check-min-healthy-warning-threshold value   Min instances check warning threshold (default 0.8)
      --slack-channel string                        #Channel / @User to post the alert (defaults to webhook configuration)
      --slack-owner string                          Comma list of owners who should be alerted on the post
      --slack-webhook string                        Slack webhook to post the alert
      --uri string                                  Marathon URI to connect
```

Example invocation would be like the following
```
$ marathon-alerts --uri http://marathon1:8080,marathon2:8080 \
                  --slack-webhook https://hooks.slack.com/services/..../ \
                  --slack-owner ashwanthkumar,slackbot
```

## Available Checks
- [x] `min-healthy` - Minimum % of Task instances should be healthy else this check is fired.

## Notifiers
- [x] Slack
- [ ] Influx
- [ ] Pager Duty
- [ ] Email

## Contribute
If you've any feature requests or issues, please open a Github issue. We accept PRs. Fork away!

## License
http://www.apache.org/licenses/LICENSE-2.0
