
#!/bin/bash

go build -o bookings cmd/web/*.go && ./bookings
./bookings -dbname=bookings -dbuser=postgres -production=false -dbpass=manoj