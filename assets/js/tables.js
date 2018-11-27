// Local
const $dataTables = $('table.table-datatable');

$dataTables.each(function (i) {

    let order = [[0, 'asc']];
    let pageLength = 100;
    let paging = true;
    let ordering = true;

    // Limit
    const limit = $(this).attr('data-limit');
    if (limit > 0) {
        pageLength = Number(limit);
    }

    // Ordering
    const noSort = $(this).attr('data-no-sort');
    if (noSort) {
        ordering = false;
    }

    // Paging
    if ($(this).find('tbody tr').length <= pageLength) {
        paging = false;
    }

    // Sort
    const tableSort = $(this).attr('data-sort');
    if (tableSort) {

        order = eval(tableSort);

    } else {

        const $column = $(this).find('thead tr th[data-sort]');
        if ($column.length > 0) {

            const index = $column.index();
            const sort = $column.attr('data-sort');

            order = [[index, sort]];
        }
    }

    // Find
    const disabled = [];
    $(this).find('thead tr th[data-disabled]').each(function (i) {
        disabled.push($(this).index());
    });

    // Init
    $(this).DataTable({
        "pageLength": pageLength,
        "order": order,
        "paging": paging,
        "ordering": ordering,
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
                $('.dt-pagination').hide();
            }
        }
    });

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
    }
});

// Local events
$dataTables.on('page.dt', function (e, settings, processing) {

    const top = $(this).prev().offset().top - 15;
    $('html, body').animate({scrollTop: top}, 200);

});

// Server side
const dtDefaultOptions = {
    "ajax": function (data, callback, settings) {

        delete data.columns;
        delete data.length;
        delete data.search.regex;

        $.ajax({
            url: $(this).attr('data-path'),
            data: data,
            success: callback,
            dataType: 'json',
            cache: $(this).attr('data-cache') !== "false"
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

        const api = this.api();
        if (api.page.info().pages <= 1) {
            $('.dt-pagination').hide();
        }

        $(".paginate_button > a").on("focus", function () {
            $(this).blur(); // Fixes scrolling to pagination on every click
        });
    }
};

// Server side events
$('table.table-datatable2').on('page.dt search.dt', function (e, settings, processing) {

    $(this).fadeTo(500, 0.3);

    if (e.type === 'page') {

        const top = $(this).prev().offset().top - 15;
        $('html, body').animate({scrollTop: top}, 500);
    }

}).on('draw.dt', function (e, settings, processing) {

    $(this).fadeTo(100, 1);
    highLightOwnedGames();

});

// Lock icon
const $lockIcon = '<i class="fa fa-lock text-muted" data-toggle="tooltip" data-placement="left" title="Private"></i>';

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
                v.createdCell($td[0], null, data, null, null); // todo, this [0] may not be needed
            }

            $td.find('[data-livestamp]').html('a few seconds ago');

            $row.append($td);
        }
    }


    $table.prepend($row);

    $table.find('tbody tr').slice(limit).remove();
}
