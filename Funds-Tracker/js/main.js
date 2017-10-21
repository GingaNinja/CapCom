$(document).ready(function() {
  console.log("hello");

  $.ajax({
      url: 'https://psd2-api.openbankproject.com/my/logins/direct',
      type: 'POST',
      
      success: function() { alert('POST completed'); },
      contentType: 'application/json',
      headers: {
        'Authorization': 'DirectLogin username="robert.xuk.x@example.com",password="5232e7",consumer_key="s2dvxvzz5v3anhkn0rzg0mjjketbvmwgoopkbzsh"'
      }
  });

}); // end of document ready
