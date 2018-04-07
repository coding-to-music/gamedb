if ($('#price-changes-page').length > 0) {

    $.fn.dataTable.ext.search.push(
        function (settings, data, dataIndex) {

            var change = Number(data[5].replace(/[^0-9\.-]+/g, ""));

            if ($('#hide-red').is(':checked') && change > 0) {
                return false;
            }

            if ($('#hide-green').is(':checked') && change < 0) {
                return false;
            }

            return true;
        }
    );

    $('#hide-red, #hide-green, #hide-owned').change(function () {

        $('#DataTables_Table_0').DataTable().draw();

    })
}
