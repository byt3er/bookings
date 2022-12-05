function Prompt() {
    let toast = function (c) {
        const{
            msg = '',
            icon = 'success',
            position = 'top-end',

        } = c;

        const Toast = Swal.mixin({
            toast: true,
            title: msg,
            position: position,
            icon: icon,
            showConfirmButton: false,
            timer: 3000,
            timerProgressBar: true,
            didOpen: (toast) => {
                toast.addEventListener('mouseenter', Swal.stopTimer)
                toast.addEventListener('mouseleave', Swal.resumeTimer)
            }
        })

        Toast.fire({})
    }

    let success = function (c) {
        const {
            msg = "",
            title = "",
            footer = "",
        } = c

        Swal.fire({
            icon: 'success',
            title: title,
            text: msg,
            footer: footer,
        })

    }

    let error = function (c) {
        const {
            msg = "",
            title = "",
            footer = "",
        } = c

        Swal.fire({
            icon: 'error',
            title: title,
            text: msg,
            footer: footer,
        })

    }

    async function custom(c) {
        const {
            icon ="",
            msg = "",
            title = "",
            showConfirmButton = true,
        } = c;

        const { value: result } = await Swal.fire({
            icon: icon,
            title: title,
            html: msg,
            backdrop: false,
            focusConfirm: false,
            showCancelButton: true,
            showConfirmButton: showConfirmButton,
            willOpen: () => {
               if(c.willOpen !== undefined){
                    c.willOpen();
               }
            },
            didOpen: () => {
                if(c.didOpen !== undefined){
                    c.didOpen()
                }
            },
            preConfirm: () => {
                return [
                    document.getElementById('start').value,
                    document.getElementById('end').value
                ]
            }
        })
        // check there is any result AND they've not click the cancel button
        // on the prompt window AND the result is not exactly equal to empty string.
        if(result){
            // if they didn't hit the cancel button, 
            // then check to see if we have any actual values.
            if(result.dismiss !== Swal.DismissReason.cancel){
                //check the result is not empty string
                if(result.value !== ""){
                    if(c.callback !== undefined){
                        c.callback(result);
                    }
                }else{
                    c.callback(false);
                }
            }else{
                c.callback(false);
            }

        }


        /*
        if (formValues) {
            Swal.fire(JSON.stringify(formValues))
        }
        */
    }

    return {
        toast: toast,
        success: success,
        error: error,
        custom: custom,
    }
}

function checkAvailabilityButton(room_id){
    document.getElementById("check-availability-button").addEventListener("click", function () {
        let html = `
        <form id="check-availability-form" action="" method="post" novalidate class="needs-validation">
            <div class="form-row">
                <div class="col">
                    <div class="form-row" id="reservation-dates-modal">
                        <div class="col">
                            <input disabled required class="form-control" type="text" name="start" id="start" placeholder="Arrival">
                        </div>
                        <div class="col">
                            <input disabled required class="form-control" type="text" name="end" id="end" placeholder="Departure">
                        </div>
                    </div>
                </div>
            </div>
        </form>
        `;
        attention.custom({
            title: 'Choose your dates',
            msg: html,
            willOpen: () => {
                const elem = document.getElementById("reservation-dates-modal");
                const rp = new DateRangePicker(elem, {
                    format: 'yyyy-mm-dd',
                    showOnFocus: true,
                    minDate: new Date(),
                })
            },
            didOpen: () => {
                document.getElementById("start").removeAttribute("disabled");
                document.getElementById("end").removeAttribute("disabled");
            },
            callback: function(result) {
                


                var data = new FormData();    
                data.append("start", document.getElementById("start").value);
                data.append("end", document.getElementById("end").value);
                data.append("csrf_token", "jHKd9Gkzd268m1XvhEX8JeWLyOkM\/euPhgz\/gYUDsDs7eQ0uLNLNJVgTjWhP4fZ\u002b4vYizwMvyE07xu\/e59Z06Q==");
                data.append("room_id",room_id);
                
                
                fetch("/search-availability-json", {
                    method: "POST",
                    body: data
                })
                    .then(response => response.json())
                    .then(data => {
                        if (data.ok){
                            attention.custom({
                                icon: 'success',
                                msg: '<p>Room is Available!</p>'
                                    +'<p><a href="/book-room?id='
                                    +data.room_id
                                    +'&sd='
                                    +data.start_date
                                    +'&ed='
                                    +data.end_date
                                    +'" class="btn btn-primary">'
                                    +'Book Now!</a></p>',
                                showConfirmButton: false,

                            })
                        }else{
                            attention.error({
                                msg: "No Availablity",
                            })
                        }
                    })
            }
        });
    })

}