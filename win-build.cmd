set DEVEL=c:\tmp\devel
set ROOT=%DEVEL%\sipswitchd
set SRC=%ROOT%\src
set GOPATH=%ROOT%

@rem pushd %SRC%
@rem hg clone https://sdr@bitbucket.org/sdr/sip_parser
@rem move sip_parser\src\*.* sip_parser
@rem popd

gofmt -s -w %SRC%

go install sip_parser
go install sipswitchd
