const $appPage = $('#app-page');

if ($appPage.length > 0) {

    const $modal = $('#news-modal');

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

    // Add hash when clicking row
    $('#news table.table').on('click', 'td', function (e) {
        history.pushState(undefined, undefined, '#news,' + $(this).closest('tr').attr('data-id'));
        showArt();
    });

    // Remove hash when closing modal
    $modal.on('hidden.bs.modal', function (e) {
        history.pushState("", document.title, "#news");
        showArt();
    });

    // News modal
    $(window).on('hashchange', showArt);
    $(document).on('draw.dt', showArt);

    // Link to dev tabs
    $(document).ready(function (e) {
        const hash = window.location.hash;
        if (hash.startsWith('#dev-')) {
            $('a.nav-link[href="#dev"]').tab('show');
            $('a.nav-link[href="' + hash + '"]').tab('show');
        }
    });

    function showArt() {

        const split = window.location.hash.split(',');

        // If the hash has a news ID
        if (split.length === 2 && (split[0] === 'news' || split[0] === '#news') && split[1]) {

            let $art = $('tr[data-id=' + split[1] + ']').find('.d-none').html();
            $art = $("<div />").html($art).text(); // Decode HTML
            $modal.find('.modal-body').html($art);
            $modal.modal('show');

        } else {
            $modal.modal('hide');
        }
    }

    // Websockets
    websocketListener('app', function (e) {

        const data = $.parseJSON(e.data);
        if (data.Data.toString() === $appPage.attr('data-id')) {
            toast(true, 'Click to refresh', 'This app has been updated', 0, 'refresh');
        }

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
                    return '<div><i class="fas fa-newspaper"></i> ' + row[1] + '</div><div class="d-none">' + row[5] + '</div>';
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
