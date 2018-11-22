if ($('#app-page').length > 0) {

    // Background
    const background = $('.container[data-bg]').attr('data-bg');
    if (background !== '') {
        $('body').css("background-image", 'url(' + background + ')');
    }

    // Fix links
    $('#news a').each(function () {

        const href = $(this).attr('href');
        if (href && !(href.startsWith('http'))) {
            $(this).attr('href', 'http://' + href);
        }
    });

    // News modal
    $(window).on('hashchange', showArt);
    $(document).on('draw.dt', showArt);

    function showArt() {
        const split = window.location.hash.split(',');
        if (split.length === 2 && (split[0] === 'news' || split[0] === '#news')) {
            const content = $('tr[data-id=' + split[1] + ']').find('.d-none').html();
            $('#news-modal .modal-body').html(content);
            $('#news-modal').modal('show');
        }
    }

    $('#news table.table').on('click', 'td', function (e) {
        window.location.hash = 'news,' + $(this).closest('tr').attr('data-id');
    });

    // News data table
    $('table.table-datatable2').DataTable($.extend(true, {}, dtDefaultOptions, {
        "order": [[2, 'desc']],
        "createdRow": function (row, data, dataIndex) {
            $(row).attr('data-id', data[0]);
        },
        "columnDefs": [
            // Title
            {
                "targets": 0,
                "render": function (data, type, row) {
                    return '<div>' + row[1] + '</div><div class="d-none">' + row[5] + '</div>';
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).addClass('article-title');
                },
                "orderable": false
            },
            // Author
            {
                "targets": 1,
                "render": function (data, type, row) {
                    return row[2];
                },
                "orderable": false
            },
            // Date
            {
                "targets": 2,
                "render": function (data, type, row) {
                    return '<span data-toggle="tooltip" data-placement="left" title="' + row[4] + '" data-livestamp="' + row[3] + '"></span>';
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).attr('nowrap', 'nowrap');
                },
                "orderable": false
            }
        ]
    }));
}
