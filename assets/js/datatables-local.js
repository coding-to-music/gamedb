$("table.table-datatable").each(function (i) {

    var order = [[0, 'asc']];

    // Find default column to sort by
    var $column = $(this).find('thead tr th[data-sort]');
    if ($column.length > 0) {

        var index = $column.index();
        var sort = $column.attr('data-sort');

        order = [[index, sort]];
    }

    // Find
    var disabled = [];
    $(this).find('thead tr th[data-disabled]').each(function (i) {
        disabled.push($(this).index());
    });

    // Init
    $(this).DataTable({
        "order": order,
        "paging": false,
        "ordering": true,
        "info": false,
        "searching": true,
        "search": {
            "smart": true
        },
        "autoWidth": false,
        "lengthChange": false,
        "stateSave": false,
        "dom": 't',
        "columnDefs": [
            {
                "targets": disabled,
                "orderable": false,
                "searchable": false
            }
        ]
    });

});

// Filter table on search box enter key
var table = $('#DataTables_Table_0');
if (table.length) {
    table = table.DataTable();

    $('input#search').keypress(function (e) {
        if (e.which === 13) {
            table.search($(this).val()).draw();
        }
    });
}

// Clear search box on escape and reset filter
$('input#search').on('keyup', function (e) {
    if ($(this).val()) {
        // var code = e.charCode || e.keyCode;
        if (e.key === "Escape") {
            $(this).val('');

            var table = $('#DataTables_Table_0');
            if (table.length) {
                table.DataTable().search($(this).val()).draw();
            }
        }
    }
});
