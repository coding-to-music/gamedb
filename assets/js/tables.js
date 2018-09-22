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

function getData(data, callback, settings, url) {

    delete data.columns;

    var trs = $('.table-datatable2 tbody tr');
    data.first = trs.first().attr('data-id');
    data.last = trs.last().attr('data-id');
    data.direction = paginationButtonClicked;

    console.log(paginationButtonClicked);

    $.ajax({
        url: url,
        data: data,
        success: function (rdata) {
            callback(rdata);
        },
        dataType: 'json',
        cache: true
    });
}

var options = {
    "processing": true,
    "serverSide": true,
    "language": {
        "processing": '<i class="fas fa-spinner fa-spin fa-3x fa-fw"></i>'
    },
    "pageLength": 100,
    "fixedHeader": true,
    "paging": true,
    "ordering": true,
    "pagingType": "full",
    "info": false,
    "searching": true,
    "search": {
        "smart": true
    },
    "autoWidth": false,
    "lengthChange": false,
    "stateSave": true,
    "dom": 'r<"dt-pagination"p>t',
    "drawCallback": function (settings, json) {
        $(".paginate_button > a").on("focus", function () {
            $(this).blur();
        });
    }
};

// Free Games Page
$('#free-games-page table.table-datatable2').DataTable($.extend(true, {}, options, {
    "ajax": function (data, callback, settings) {
        getData(data, callback, settings, "/free-games/ajax");
    },
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

var paginationButtonClicked;

$(document).on('mousedown', '.dt-pagination ul.pagination li a', function (e) {
    paginationButtonClicked = $(this).html().toLowerCase();
});
