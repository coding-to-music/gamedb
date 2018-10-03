
// Local datatable
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


// Server side datatable events
$('table.table-datatable2').on('page.dt', function (e, settings, processing) {

    $(this).fadeTo(500, 0.3);

    var top = $(this).closest('.card').offset().top + 5;
    $('html, body').animate({scrollTop: top}, 500);

}).on('draw.dt', function (e, settings, processing) {

    $(this).fadeTo(100, 1);

});

// Defaults
var $lockIcon = '<i class="fa fa-lock text-muted" data-toggle="tooltip" data-placement="left" title="Private"></i>';

var dtDefaultOptions = {
    "ajax": function (data, callback, settings) {

        delete data.columns;
        delete data.length;
        delete data.search.regex;

        $.ajax({
            url: $(this).attr('data-path'),
            data: data,
            success: callback,
            dataType: 'json',
            cache: true
        });

    },
    "processing": false,
    "serverSide": true,
    "pageLength": 100,
    "fixedHeader": true,
    "paging": true,
    "ordering": true,
    "info": false,
    "searching": true,
    "autoWidth": false,
    "lengthChange": false,
    "stateSave": false,
    "orderMulti": false,
    "dom": '<"dt-pagination"p>t<"dt-pagination"p>',
    "language": {
        "processing": '<i class="fas fa-spinner fa-spin fa-3x fa-fw"></i>'
    },
    "drawCallback": function (settings, json) {
        $(".paginate_button > a").on("focus", function () {
            $(this).blur(); // Fixes scrolling to pagination on every click
        });
    }
};

function addDataTablesRow(columnDefs, data, limit, $table) {

    var $row = $('<tr />');

    for (var i in columnDefs) {
        if (columnDefs.hasOwnProperty(i)) {

            var value = data[i];

            if ('render' in columnDefs[i]) {
                value = columnDefs[i].render(null, null, data);
            }

            var $td = $('<td />').html(value);

            if ('createdCell' in columnDefs[i]) {
                columnDefs[i].createdCell($td[0], null, data, null, null); // todo, this [0] may not be needed
            }

            $row.append($td);
        }
    }

    $table.prepend($row);

    $table.find('tbody tr').slice(limit).remove();
}
