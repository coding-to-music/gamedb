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

var options = {
    "processing": true,
    "serverSide": true,
    "ajax": {
        "cache": true,
        "data": function (d) {
            delete d.columns;
        }
    },
    "fixedHeader": true,
    "paging": true,
    "ordering": true,
    "info": false,
    "searching": true,
    "search": {
        "smart": true
    },
    "autoWidth": false,
    "lengthChange": false,
    "stateSave": false,
    //"dom": 't',
    "drawCallback": function (settings, json) {
        $(".paginate_button > a").on("focus", function () {
            $(this).blur();
        });
    }
};

// Init
$('#free-games-page table.table-datatable2').DataTable($.extend(true, {}, options, {
    "ajax": {
        "url": "/free-games/ajax"
    },
    "order": [[1, 'desc']],
    "createdRow": function (row, data, dataIndex) {
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

// Search
var table = $('#DataTables_Table_0');
if (table.length) {
    table = table.DataTable();

    $('input#search').keyup(function () {
        table.search($(this).val()).draw();
    });
}

// Layout
$('.dataTables_wrapper').removeClass('container-fluid');

// Stop scrolling to bottom bug
$("table.table-datatable, table.table-datatable2").on('page.dt', function (e, settings) {
    $('html, body').animate({scrollTop: 0}, 'fast');
});
