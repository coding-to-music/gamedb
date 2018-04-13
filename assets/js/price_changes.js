if ($('#price-changes-page').length > 0) {

    var $hideRed = $('#hide-red');
    var $hideGreen = $('#hide-green');
    var $hideApps = $('#hide-apps');
    var $hidePackages = $('#hide-packages');
    var $hideOwned = $('#hide-owned');

    $.fn.dataTable.ext.search.push(
        function (settings, searchData, index, rowData, counter) {

            var change = Number(searchData[5].replace(/[^0-9\.-]+/g, ""));

            if ($hideRed.is(':checked') && change > 0) {
                return false;
            }

            if ($hideGreen.is(':checked') && change < 0) {
                return false;
            }

            if ($hideApps.is(':checked')) {

                var appID = table
                    .row(index)         //get the row to evaluate
                    .nodes()                //extract the HTML - node() does not support to$
                    .to$()                  //get as jQuery object
                    // .find('td[data-label]') //find column with data-label
                    // .data('label');         //get the value of data-label
                    .attr('data-app-id');

                if (appID > 0) {
                    return false;
                }
            }

            if ($hidePackages.is(':checked')) {

                var packageID = table
                    .row(index)         //get the row to evaluate
                    .nodes()                //extract the HTML - node() does not support to$
                    .to$()                  //get as jQuery object
                    // .find('td[data-label]') //find column with data-label
                    // .data('label');         //get the value of data-label
                    .attr('data-package-id');

                if (packageID > 0) {
                    return false;
                }
            }

            return true;
        }
    );

    $('#hide-red, #hide-green, #hide-apps, #hide-packages, #hide-owned').change(function () {

        $('#DataTables_Table_0').DataTable().draw();

    })
}
