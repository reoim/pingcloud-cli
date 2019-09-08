# pingcloud-cli
![pingcloud-cli gcp](/assets/images/gcp-all.png)

Pingcloud-cli is CLI application to check latency and http trace from **AWS, GCP and Azure** regions.

This is a personal toy project. 

Inspired by [GCPing](https://github.com/GoogleCloudPlatform/gcping), [httpstat-1](https://github.com/reoim/httpstat-1) and [AzureSpeed](https://github.com/blrchen/AzureSpeed).


## Capabilities

1.  Send HTTP(S) request to regions of AWS, GCP and Azure. Calculate and report response time.
2.  Print httpstat of region(s).
3.  List region code and region name of AWS, GCP and Azure.


```
Thank you for using pingcloud-cli.
pingcloud-cli is command line tools to check latency and http trace from AWS, GCP and Azure regions.
You can download codes from https://github.com/reoim/pingcloud-cli
Any feedback is welcome. (And stars..)

Usage:
  pingcloud-cli [command]

Available Commands:
  aws         Check latencies of AWS regions.
  azure       Check latencies of Azure regions.
  gcp         Check latencies of GCP regions.
  help        Help about any command
  version     Print version of pingcloud-cli

Flags:
  -h, --help   help for pingcloud-cli

Use "pingcloud-cli [command] --help" for more information about a command.
```

## Prerequisite
1.  GO installed.
2.  $GOPATH and $GOBIN are properly set.



## Installation
1.  Run `git clone https://github.com/reoim/pingcloud-cli.git`
2.  Run `GO111MODULE=on go mod vendor`
3.  Run `go build -v`
4.  Run `go install`


## Usage
### GCP (Google Cloud Platform)
1.  `pingcloud-cli gcp`

    Ping all GCP regions

    ![pingcloud-cli gcp](/assets/images/gcp.png)
2.  `pingcloud-cli gcp [region code]`

    Print httpstat of specific regions. You can append multiple region codes to the end of command. (Seperate each region with space)

    ![pingcloud-cli gcp region](/assets/images/gcp-region.png)
 

### AWS
1.  `pingcloud-cli aws`

    Ping all AWS regions

    ![pingcloud-cli aws](/assets/images/aws.png)
2.  `pingcloud-cli aws [region code]`

    Print httpstat of specific regions. You can append multiple region codes to the end of command. (Seperate each region with space)

    ![pingcloud-cli aws region](/assets/images/aws-region.png)

### Azure
1.  `pingcloud-cli azure`

    Ping all Azure regions

    ![pingcloud-cli azure](/assets/images/azure.png)
2.  `pingcloud-cli azure [region code]`

    Print httpstat of specific regions. You can append multiple region codes to the end of command. (Seperate each region with space)

    ![pingcloud-cli azure region](/assets/images/azure-region.png)

### Flags
* `-l or --list`
List all region codes and region names of the cloud provider. Add -l or --list flag after command.

![pingcloud-cli gcp -l](/assets/images/gcp-list.png) 



## Notes
Instances for latency test are not maintained by me. 

Endpoints for AWS are from [EC2 Reachability Test](http://ec2-reachability.amazonaws.com/).

For GCP, [GCP ping](http://www.gcping.com/)

And for Azure, [AzureSpeed](https://github.com/blrchen/AzureSpeed).

Endpoints from Azure have domain name and uses https but AWS, GCP endpoints are static and use http. 

So latencies from Azure are relatively high compare to AWS and GCP because it needs domain loockup and tls handshaking.

### 2019-09-08
Changed Azure test endpoints https -> http.

No more TLS handshaking time.
