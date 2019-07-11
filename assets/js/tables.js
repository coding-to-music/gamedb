// Local
const $dataTables = $('table.table-datatable');
const $dataTables2 = $('table.table-datatable2');
let dataTables = [];

const $lockIcon = '<i class="fa fa-lock text-muted" data-toggle="tooltip" data-placement="left" title="Private"></i>';

$dataTables.each(function (i) {

    // Find
    const disabled = [];
    $(this).find('thead tr th[data-disabled]').each(function (i) {
        disabled.push($(this).index());
    });

    // Init
    const dt = $(this).DataTable({
        "pageLength": 100,
        "paging": true,
        "ordering": true,
        "fixedHeader": true,
        "info": false,
        "searching": true,
        "search": {
            "smart": true
        },
        "autoWidth": false,
        "lengthChange": false,
        "stateSave": false,
        "dom": '<"dt-pagination"p>t<"dt-pagination"p>',
        "columnDefs": [
            {
                "targets": disabled,
                "orderable": false
            }
        ],
        "drawCallback": function (settings, json) {

            const api = this.api();
            if (api.page.info().pages <= 1) {
                $(this).parent().find('.dt-pagination').hide();
            }
        },
        "initComplete": function (settings, json) {

            $('table.table-datatable').on('order.dt', function (e, settings, processing) {

                $('#live-badge').trigger('click');

            });
        }
    });

    dataTables.push(dt);

});

// Local search
const $searchField = $('input#search');
$searchField.on('keyup', function (e) {
    $dataTables.DataTable().search($(this).val()).draw();
});

$searchField.on('keyup', function (e) {
    if ($(this).val() && e.key === "Escape") {
        $(this).val('');
        $dataTables.DataTable().search($(this).val()).draw();
        $dataTables2.DataTable().search($(this).val()).draw();
    }
});

// Local events
$dataTables.on('page.dt', function (e, settings, processing) {

    let padding = 15;

    if ($('.fixedHeader-floating').length > 0) {
        padding = padding + 48;
    }

    $('html, body').animate({
        scrollTop: $(this).prev().offset().top - padding
    }, 200);
});

$dataTables.on('draw.dt', function (e, settings) {

    // fixBrokenImages();
    observeLazyImages('tr img[data-lazy]');
});

function getPagingType() {

    if (user.userLevel === "0") {
        return 'simple_numbers'
    } else {
        return 'simple_numbers'
    }
}

// Server side
const dtDefaultOptions = {
    "ajax": function (data, callback, settings, $ctx = null) {

        let $this = $(this);
        if ($ctx) {
            $this = $ctx;
        }

        $.ajax({
            url: function () {
                const path = $this.attr('data-path');
                if (!path && user.isLocal) {
                    console.log('Table data-path not set');
                }
                return path;
            }(),
            data: data,
            success: function (data, textStatus, jqXHR) {

                callback(data, textStatus, jqXHR);

                // noinspection JSUnresolvedVariable
                if (data.limited) {
                    const bold = $('li.paginate_button.page-item.next.disabled').length > 0 ? 'font-weight-bold' : '';
                    const donate = $('<li class="donate"><small><a href="/donate"><i class="fas fa-heart text-danger"></i> <span class="' + bold + '">See more!</span></a></small></li>');
                    $this.parent().find('.dt-pagination ul.pagination').append(donate);
                }

            },
            error: function (jqXHR, textStatus, errorThrown) {

                const data = {
                    "draw": "1",
                    "recordsTotal": "0",
                    "recordsFiltered": "0",
                    "data": [],
                    "limited": false
                };

                callback(data, textStatus, null);
            },
            dataType: 'json',
            cache: $this.attr('data-cache') !== "false"
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
    "orderMulti": false,
    "pagingType": getPagingType(),
    "dom": '<"dt-pagination"p>t<"dt-pagination"p>r',
    "language": {
        "processing": '<i class="fas fa-spinner fa-spin fa-3x fa-fw"></i>',
        "paginate": {
            "next": '<i class="fas fa-chevron-right"></i>',
            "previous": '<i class="fas fa-chevron-left"></i>',
        },
    },
    "drawCallback": function (settings, json) {

        const api = this.api();
        if (api.page.info().pages <= 1) {
            $(this).parent().find('.dt-pagination').hide();
        }

        $(".paginate_button > a").on("focus", function () {
            $(this).blur(); // Fixes scrolling to pagination on every click
        });
    },
    "initComplete": function (settings, json) {

        $dataTables2.on('order.dt', function (e, settings, processing) {

            $('#live-badge').trigger('click');

        });
    }
};

// Server side events
$dataTables2.filter(':not(.table-no-fade)').on('page.dt search.dt', function (e, settings, processing) {

    $(this).fadeTo(500, 0.3);

}).on('draw.dt', function (e, settings, processing) {

    $(this).fadeTo(100, 1);
});

$dataTables2.on('page.dt', function (e, settings, processing) {

    $('html, body').animate({
        scrollTop: $(this).prev().offset().top - 15
    }, 200);
});

$dataTables2.on('draw.dt', function (e, settings, processing) {

    highLightOwnedGames();
    observeLazyImages('tr img[data-lazy]');
    fixBrokenImages();
});

//
function addDataTablesRow(options, data, limit, $table) {

    let $row = $('<tr class="fade-green" />');
    options.createdRow($row[0], data, null);

    if (isIterable(options.columnDefs)) {
        for (const v of options.columnDefs) {

            let value = data[v];

            if ('render' in v) {
                value = v.render(null, null, data);
            }

            const $td = $('<td />').html(value);

            if ('createdCell' in v) {
                v.createdCell($td, null, data, null, null);
            }

            $td.find('[data-livestamp]').html('a few seconds ago');

            $row.append($td);
        }
    }


    $table.prepend($row);

    $table.find('tbody tr').slice(limit).remove();
}
