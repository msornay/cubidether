$(function() {
    //hang on event of form with id=myform
    $("#form").submit(function(e) {
      	var coinbase = $('#form').find('input[name="coinbase"]').val();
  		alert(address)
    	console.log(address)

		$.ajax({
		    contentType: 'application/json',
		    data: {
		        "coinbase": coinbase
		    },
		    dataType: 'json',
		    success: function(data){
		    	//parse response
		    },
		    error: function(){
		    },
		    type: 'POST',
		    url: '/'
		});

    });
});

