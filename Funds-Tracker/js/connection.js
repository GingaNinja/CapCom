$(document).ready(function() {

  $.getJSON("http://localhost:8080/api/getnexttransaction", function(data) {
    var ribbonViewModel;

    ribbonViewModel = ko.mapping.fromJS(data);
        ko.applyBindings(ribbonViewModel);
        // console.log(this.data);
    // function AppViewModel() {
    //
    //   this.description = ko.observable("Bert");
    //   this.amount = ko.observable("Bertington");
    // }

    // Activates knockout.js
    // ko.applyBindings(new AppViewModel());

  }).done(function() {
    console.log('Ajax definitely finished');
  }); // end of done
}); // end of document ready
