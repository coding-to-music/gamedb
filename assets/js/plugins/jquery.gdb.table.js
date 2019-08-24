// the semi-colon before function invocation is a safety net against concatenated
// scripts and/or other plugins which may not be closed properly.
;(function ($, window, document, undefined) {

    "use strict";

    // Create the defaults once
    var pluginName = "gdbTable";
    var defaults = {
        fadeOnLoad: true,
        cache: true,
        tableOptions: {},
    };
    var localDefaults = {
        "search": {
            "smart": true
        },
    };
    var remoteDefaults = {
        "ajax": function (data, callback, settings) {

            $.ajax({
                url: function () {
                    const path = $(this).attr('data-path');
                    if (!path && user.log) {
                        console.log('Table data-path not set');
                    }
                    return path;
                }(),
                data: data,
                success: callback,
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
                cache: $(this).attr('data-cache') !== "false"
            });
        },
        "processing": true,
        "serverSide": true,
        "orderMulti": false,
    };

    // The actual plugin constructor
    function Plugin(element, options) {

        if (options === undefined) {
            options = {}
        }

        if (options.tableOptions === undefined) {
            options.tableOptions = {};
        }

        options.isAjax = function () {
            return this.tableOptions.columnDefs !== undefined
        }

        var tableOptions = {
            "autoWidth": false,
            "dom": '<"dt-pagination"p>t<"dt-pagination"p>r',
            "fixedHeader": true,
            "info": false,
            "language": {
                "processing": '<i class="fas fa-spinner fa-spin fa-3x fa-fw"></i>',
                "paginate": {
                    "next": '<i class="fas fa-chevron-right"></i>',
                    "previous": '<i class="fas fa-chevron-left"></i>',
                },
            },
            "lengthChange": false,
            "ordering": true,
            "pageLength": 100,
            "paging": true,
            "pagingType": 'simple_numbers',
            "searching": true,
            "stateSave": false,
        }

        if (!options.isAjax()) {
            tableOptions.columnDefs = [
                {
                    "orderable": false,
                    "targets": $(element).find('thead tr th[data-disabled]').map(function () {
                        return $(this).index();
                    }).get(),
                }
            ]
        }

        this.settings = $.extend(true, {}, {tableOptions: tableOptions}, {tableOptions: (options.isAjax() ? remoteDefaults : localDefaults)}, options);
        this.element = element;
        this._defaults = defaults;
        this._name = pluginName;

        this.init();
    }

    $.extend(Plugin.prototype, {
        init: function () {

            const dt = $(this.element).DataTable(this.settings.tableOptions);

            this.scrollOnPaginate(dt);
            this.hideEmptyPagination(dt);

            if (this.settings.isAjax()) {

                if (this.settings.fadeOnLoad) {
                    this.fadeOnLoad(dt);
                }
            }

            // Keep track of tables
            if (window.gdbTables === undefined) {
                window.gdbTables = [];
            }
            window.gdbTables.push(dt);

            // Fixes scrolling to pagination on every click
            $(".paginate_button > a").one("focus", function () {
                $(this).blur();
            });

            // Fixes hidden fixed header tables
            $('a[data-toggle="tab"]').one('shown.bs.tab', function (e) {
                $.each(dataTables, function (index, value) {
                    value.fixedHeader.adjust();
                });
            });
        },
        fixImages: function (dt) {
            highLightOwnedGames();
            observeLazyImages('tr img[data-lazy]');
            fixBrokenImages();
        },
        addDonateButton: function (dt) {

            console.log(json);
            dt.on('xhr.dt', function (e, settings, json, xhr) {
                if (json.limited) {
                    const bold = $('li.paginate_button.page-item.next.disabled').length > 0 ? 'font-weight-bold' : '';
                    const donate = $('<li class="donate"><small><a href="/donate"><i class="fas fa-heart text-danger"></i> <span class="' + bold + '">See more!</span></a></small></li>');
                    $(this).parent().find('.dt-pagination ul.pagination').append(donate);
                }
            });
        },
        fadeOnLoad: function (dt) {
            dt.on('page.dt search.dt', function (e, settings, processing) {

                $(this).fadeTo(500, 0.3);

            }).on('draw.dt', function (e, settings, processing) {

                $(this).fadeTo(100, 1);
            });
        },
        hideEmptyPagination: function (dt) {
            dt.on('draw.dt', function (e, settings, processing) {

                if (dt.page.info().pages <= 1) {
                    $(this).parent().find('.dt-pagination').hide();
                } else {
                    $(this).parent().find('.dt-pagination').show();
                }
            });
        },
        scrollOnPaginate: function (dt) {
            dt.on('page.dt', function (e, settings, processing) {

                let padding = 15;

                if ($('.fixedHeader-floating').length > 0) {
                    padding = padding + 48;
                }

                $('html, body').animate({
                    scrollTop: $(this).prev().offset().top - padding
                }, 200);
            });
        },
    });

    $.fn[pluginName] = function (options) {
        return this.each(function () {
            if (!$.data(this, "plugin_" + pluginName)) {
                $.data(this, "plugin_" + pluginName, new Plugin(this, options));
            }
        });
    };

})(jQuery, window, document);