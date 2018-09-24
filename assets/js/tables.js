// https://datatables.net/reference/option/

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

var defaultOptions = {
    "ajax": function (data, callback, settings) {

        delete data.columns;
        delete data.length;

        $.ajax({
            url: $(this).attr('data-path'),
            data: data,
            success: function (rdata) {
                callback(rdata);
            },
            dataType: 'json',
            cache: true
        });

    },
    "processing": true,
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
    "dom": 'r<"dt-pagination"p>t',
    "language": {
        "processing": '<i class="fas fa-spinner fa-spin fa-3x fa-fw"></i>'
    },
    "drawCallback": function (settings, json) {
        $(".paginate_button > a").on("focus", function () {
            $(this).blur(); // Fixes scrolling to pagination on every click
        });
    }
};

// Free Games Page
$('#free-games-page table.table-datatable2').DataTable($.extend(true, {}, defaultOptions, {
    "order": [[1, 'desc']],
    "createdRow": function (row, data, dataIndex) {
        $(row).attr('data-id', data[0]);
        $(row).attr('data-link', data[7]);
    },
    "columnDefs": [
        {
            "targets": 0,
            "render": function (data, type, row) {
                return '<img src="' + row[2] + '" class="rounded square"><span>' + row[1] + '</span>';
            },
            "createdCell": function (td, cellData, rowData, row, col) {
                $(td).addClass('img')
            }
        },
        {
            "targets": 1,
            "render": function (data, type, row) {
                return row[3] + '%';
            }
        },
        {
            "targets": 2,
            "render": function (data, type, row) {
                return row[4];
            }
        },
        {
            "targets": 3,
            "render": function (data, type, row) {
                return row[5];
            },
            "createdCell": function (td, cellData, rowData, row, col) {
                $(td).addClass('platforms platforms-align')
            },
            "orderable": false,
            "searchable": false
        },
        {
            "targets": 4,
            "render": function (data, type, row) {
                return '<a href="' + row[6] + '">Install</a>';
            },
            "orderable": false,
            "searchable": false
        }
    ]
}));

// Player Apps Tab
$('#player-page #games table.table-datatable2').DataTable($.extend(true, {}, defaultOptions, {
    "order": [[2, 'desc']],
    "createdRow": function (row, data, dataIndex) {
        $(row).attr('data-id', data[0]);
        $(row).attr('data-link', '/games/' + data[0]);
    },
    "columnDefs": [
        {
            "targets": 0,
            "render": function (data, type, row) {
                return '<img src="' + row[2] + '" class="avatar rounded square"><span>' + row[1] + '</span>';
            },
            "createdCell": function (td, cellData, rowData, row, col) {
                $(td).addClass('img')
            }
        },
        {
            "targets": 1,
            "render": function (data, type, row) {
                return '$' + row[5];
            },
            "createdCell": function (td, cellData, rowData, row, col) {
                $(td).attr('data-sort', row[5])
            }
        },
        {
            "targets": 2,
            "render": function (data, type, row) {
                return row[4];
            },
            "createdCell": function (td, cellData, rowData, row, col) {
                $(td).attr('nowrap', 'nowrap').attr('data-sort', row[3])
            }
        },
        {
            "targets": 3,
            "render": function (data, type, row) {
                return '$' + row[6];
            }
        }
    ]
}));

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
